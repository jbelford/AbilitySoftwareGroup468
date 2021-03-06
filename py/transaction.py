import sys

import functools

sys.path.append('gen-py')

import time
import logging
from concurrent.futures import ThreadPoolExecutor, Future
from threading import Thread
from multiprocessing import Lock
from CmdType import Cmd

from DistQueue import DistQueue
from cache import Cache
from utils import process_error, _executor
from Service import Service
from transactionRPC import Transaction
from auditserver import AuditServer
from database import dbserver
from shared.ttypes import Response, Command, QuoteData, PendingTxn, Trigger
from databaseRPC.ttypes import DBResponse


@Service(thrift_class=Transaction, port=44421)
class transactionserver(object):

	def __init__(self, use_rpc=False, server=False, queue_to_use=0):
		self._database = dbserver(use_rpc=use_rpc, server=False)
		self._audit = AuditServer(use_rpc=use_rpc, server=False)
		self._cache = Cache(use_rpc=use_rpc, server=False, mock=True)
		self.queue = DistQueue(use_rpc=use_rpc, server=False)
		self.executor = ThreadPoolExecutor(max_workers=16)
		
		for q_i in range(1):
			worker = Thread(target=functools.partial(self.get_work, q_i))
			worker.start()
		
		
	def get_work(self, q_i):
		
		lock = Lock()
		
		while True:
			# Could be updated so a single transaction server can service multiple queues.
			lock.acquire()
			cmd = self.queue.GetItem(q_i)
			lock.release()
			if cmd.C_type < 0:
				logging.debug("Found no work.")
				continue
			
			fn = None
			
			logging.debug("Got Work: " + str(cmd.TransactionID))
			
			# Choose which function to run.
			if cmd.C_type == Cmd.ADD.value:
				fn = self.ADD
				
			elif cmd.C_type == Cmd.QUOTE.value:
				fn = self.QUOTE
				
			elif cmd.C_type == Cmd.BUY.value:
				fn = self.BUY
				
			elif cmd.C_type == Cmd.COMMIT_BUY.value:
				fn = self.COMMIT_SELL
				
			elif cmd.C_type == Cmd.CANCEL_BUY.value:
				fn = self.CANCEL_BUY
				
			elif cmd.C_type == Cmd.SELL.value:
				fn = self.SELL
			
			elif cmd.C_type == Cmd.COMMIT_SELL.value:
				fn = self.COMMIT_SELL
			
			elif cmd.C_type == Cmd.CANCEL_SELL.value:
				fn = self.CANCEL_SELL
				
			elif cmd.C_type == Cmd.SET_BUY_AMOUNT.value:
				fn = self.SET_BUY_AMOUNT
				
			elif cmd.C_type == Cmd.CANCEL_SET_BUY.value:
				fn = self.CANCEL_SET_BUY
				
			elif cmd.C_type == Cmd.SET_BUY_TRIGGER.value:
				fn = self.SET_BUY_TRIGGER
			
			elif cmd.C_type == Cmd.SET_SELL_AMOUNT.value:
				fn = self.SET_SELL_AMOUNT
			
			elif cmd.C_type == Cmd.SET_SELL_TRIGGER.value:
				fn = self.SET_SELL_TRIGGER
				
			elif cmd.C_type == Cmd.CANCEL_SET_SELL.value:
				fn = self.CANCEL_SET_SELL
			
			elif cmd.C_type == Cmd.DUMPLOG.value:
				fn = self.DUMPLOG
			
			assert fn is not None, "No Function Set!"
			
			self.executor.submit(fn, cmd) \
				.add_done_callback(
				functools.partial(self.mark_done, q_i, cmd))
			
			
	def mark_done(self, queue_num: int, cmd:Command, fn:Future):
		if fn.cancelled():
			logging.error("Failed Task: " + str(cmd.TransactionID))
		elif fn.done():
			logging.debug("Finished Task: " + str(cmd.TransactionID))
			res = fn.result()
			# This marks the transaction as completed, and stores the result internally.
			self.queue.MarkComplete(queue_num, cmd, res)
			
	def error(self, cmd, msg):
		process_error(self._audit, cmd, msg)
		return Response(Success=False, Message=msg)

	def ADD(self, cmd: Command):
		resp = self._database.AddUserMoney(userId=cmd.UserId, amount=cmd.Amount)
		if resp.error is not None:
			return self.error(cmd, "Failed to create and/or add money to account")
		_executor.submit(self._audit.AccountTransaction, *(cmd.UserId, cmd.Amount, "add", cmd.TransactionID, ))

		return Response(Success=True)

	def QUOTE(self, cmd: Command):
		resp: QuoteData = self._cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)

		if resp.error is not None:
			return self.error(cmd, "Quote server failed to respond with quote")

		return Response(Success=True, Quote=resp.Quote, Stock=resp.Symbol)

	def BUY(self, cmd: Command):
		resp: DBResponse = self._database.GetUser(userId=cmd.UserId)
		if resp.error is None or resp.user is None or resp.user.Reserved is None:
			logging.debug("User not found.")
			return self.error(cmd, "The user " + str(cmd.UserId) + " does not exist")

		user_reserved = resp.user.Reserved
		user_balance = resp.user.Balance
		if user_balance - user_reserved < cmd.Amount:
			return self.error(cmd, "Specified amount is greater than can afford.")

		quote: QuoteData = self._cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
		if quote.error is not None:
			return self.error(cmd, "Failed to get quote for that stock: BUY")

		shares = cmd.Amount // quote.Quote
		if shares <= 0:
			return self.error(cmd, "Specified amount is not enough to purchase any shares")
		cost = int(shares) * quote.Quote
		expiry = time.time() + 60

		pending = PendingTxn(UserId=cmd.UserId, Type="BUY", Price=cost,
		                     Shares=shares, Reserved=cmd.Amount, Stock=quote.Symbol, Expiry=expiry)

		self._database.PushPendingTxn(pending)
		return Response(Success=True, ReqAmount=cmd.Amount, RealAmount=cost, Shares=shares, Expiration=expiry)

	def COMMIT_BUY(self, cmd: Command):
		buy = self._database.PopPendingTxn(cmd.UserId, "BUY")
		if buy is None:
			return self.error(cmd, "There are no pending transactions.")

		resp:DBResponse = self._database.ProcessTxn(buy, True)
		if resp.error is not None:
			return self.error(cmd, "User can no longer afford this purchase.")

		log = _executor.submit(self._audit.AccountTransaction, *(cmd.UserId, buy.Reserved, "remove", cmd.TransactionID, ))
		# TODO:// Log to database.
		log.result()

		return Response(Success=True, Stock=buy.Stock, Shares=buy.Shares, Paid=buy.Price)

	def CANCEL_BUY(self, cmd):
		buy = self._database.PopPendingTxn(cmd.UserId, "BUY")
		if buy.error is None:
			return self.error(cmd, "There is no buy to cancel.")

		return Response(Success=True, Stock=buy.Stock, Shares=buy.Shares)

	def SELL(self, cmd: Command):
		resp = self._database.GetUser(cmd.UserId)
		if resp.error is not None or resp.user == {} \
				or resp.user is None or resp.user.stock == {}\
				or resp.user.stock is None:
			return self.error(cmd, "The user " + str(cmd.UserId) + " does not exist.")
		elif cmd.StockSymbol not in resp.user.stock.keys():
			return self.error(cmd, "The user " + str(cmd.UserId) + " does not own any stocks.")
		elif resp.user.stock[cmd.StockSymbol]["real"] == 0:
			return self.error(cmd, "User does not own any shares for that stock")

		quote: QuoteData = self._cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
		if quote.error is not None:
			return self.error(cmd, "Failed to get quote for that stock.")

		actual_shares = cmd.Amount // quote.Quote
		shares = actual_shares
		if shares <= 0:
			return self.error(cmd, "A single share is worth more than specified amount.")
		elif cmd.StockSymbol not in resp.user.stock.keys():
			return self.error(cmd, "User does not own any stocks.")
		elif resp.user.stock[cmd.StockSymbol]["real"] < shares:
			shares = resp.user.stock[cmd.StockSymbol]["real"]

		sell_for = shares * quote.Quote
		expiry = time.time() + 60
		pending = PendingTxn(UserId=cmd.UserId, Type="SELL", Price=sell_for,
		                     Shares=shares, Stock=quote.Symbol, Expiry=expiry)
		self._database.PushPendingTxn(pending)

		return Response(Success=True, ReqAmount=cmd.Amount, RealAmount=actual_shares * quote.Quote,
		                Shares=actual_shares, SharesAfford=shares, AffordAmount=sell_for, Expiration=expiry)


	def COMMIT_SELL(self, cmd: Command):
		sell:PendingTxn = self._database.PopPendingTxn(cmd.UserId, "SELL")
		if sell.error is None:
			return self.error(cmd, "There are no pending transactions.")

		resp:DBResponse = self._database.ProcessTxn(sell, True)
		if resp.error is not None:
			return self.error(cmd, "User no longer has the correct number of shares.")
		
		_executor.submit(self._audit.AccountTransaction, *(cmd.UserId, cmd.Amount, "add", cmd.TransactionID, ))

		# TODO:// Log into database?... (trying to just write to file, see how it works)

		return Response(Success=True, Stock=sell.Stock, Shares=sell.Shares, Received=sell.Price)

	def CANCEL_SELL(self, cmd: Command):
		sell = self._database.PopPendingTxn(cmd.UserId, "SELL")
		if sell.error is None:
			return self.error(cmd, "There is no sell to cancel")

		return Response(Success=True, Stock=sell.Stock, Shares=sell.Shares)

	def SET_BUY_AMOUNT(self, cmd: Command):
		resp = self._database.GetUser(cmd.UserId)
		if resp.error is not None or resp.user is None or resp.user.Reserved is None:
			return self.error(cmd, "The user does not exist.")

		user_reserved = resp.user.Reserved
		user_balance = resp.user.Balance
		if user_balance - user_reserved < cmd.Amount:
			return self.error(cmd, "Not enough funds.")

		# Skipping getting a quote here, hope we can later...

		trigger = Trigger(UserId=cmd.UserId, Stock=cmd.StockSymbol,
		                  TransactionID=cmd.TransactionID, Type="BUY", Amount=cmd.Amount, When=0)

		reserved = self._database.ReserveMoney(cmd.UserId, cmd.Amount)
		if reserved.error is not None:
			return self.error(cmd, "Failed to reserve even though we should have.")

		# TODO:// Set Trigger and unreserve money if fails...
		trig = self._database.AddNewTrigger(trigger)
		if trig.error is not None:
			self._database.UnreserveMoney(cmd.UserId, cmd.Amount)
			return self.error(cmd, "Failed to set trigger in DB.")

		_executor.submit(self._audit.AccountTransaction, *(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID, ))
		return Response(Success=True)


	def CANCEL_SET_BUY(self, cmd: Command):

		resp = self._database.GetUser(cmd.UserId)
		if resp.error is not None:
			return self.error(cmd, "The user does not exist.")

		# TODO:// Cancel Triggers and Unreserve money...
		trig: Trigger = self._database.CancelTrigger(cmd.UserId, cmd.StockSymbol, "BUY")

		if trig.error is not None:
			return self.error(cmd, "No buy trigger to cancel.")
		
		resp = self._database.UnreserveMoney(cmd.UserId, trig.Amount)
		
		if resp.error is not None:
			logging.error(resp.error)
			return self.error(cmd, "Internal server error.")

		_executor.submit(self._audit.AccountTransaction, *(cmd.Amount, trig.Amount, "unreserve", cmd.TransactionID, ))

		return Response(Success=True, Stock=cmd.StockSymbol)

	def SET_BUY_TRIGGER(self, cmd: Command):
		resp = self._database.GetUser(cmd.UserId)
		if resp is not None:
			return self.error(cmd, "The user does not exist.")

		# TODO:// Triggers..
		trig: Trigger = self._database.GetTrigger(cmd.UserId, cmd.StockSymbol, "BUY")
		if trig.error is not None:
			return self.error(cmd, "User ust set buy amount first.")

		trig.When = cmd.Amount
		# Update that trigger...
		resp = self._database.AddNewTrigger(trig)
		if resp.error is not None:
			return self.error(cmd, "Internal error during operation.")

		return Response(Success=True)

	def SET_SELL_AMOUNT(self, cmd: Command):
		resp = self._database.GetUser(cmd.UserId)
		if resp.error is not None:
			return self.error(cmd, "The user does not exist.")

		reserved = self._database.GetReservedShares(cmd.UserId, cmd.StockSymbol)
		
		if resp.error is not None or resp.user == {} \
				or resp.user is None or resp.user.stock == {}\
				or resp.user.stock is None:
			return self.error(cmd, "The user " + str(cmd.UserId) + " does not exist.")
		elif cmd.StockSymbol not in resp.user.stock.keys():
			return self.error(cmd, "The user " + str(cmd.UserId) + " does not own any stocks.")
		elif reserved < 0 or resp.user.stock[cmd.StockSymbol]["real"] == 0:
			return self.error(cmd, "User does not own any shares for that stock")
		
		real_stocks = resp.user.stock[cmd.StockSymbol]["real"] - reserved
		if real_stocks <= 0:
			return self.error(cmd, "The user does not have any stocks.")

		quote = self._cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
		if quote.error is not None:
			return self.error(cmd, "Failed to get quote for that stock.")

		reserved_shares = cmd.Amount // quote.Quote
		if reserved_shares > real_stocks:
			reserved_shares = real_stocks

		trigger = Trigger(UserId=cmd.UserId, Type="SELL", TransactionID=cmd.TransactionID,
		                  Shares=reserved_shares, Stock=cmd.StockSymbol, Amount=cmd.Amount, When=0)

		# TODO:// Set the trigger...
		resp = self._database.AddNewTrigger(trigger)
		if resp.error is not None:
			return self.error(cmd, "Failed to set sell amount.")
		self._database.ReserveShares(cmd.UserId, cmd.StockSymbol, reserved_shares)
		_executor.submit(self._audit.AccountTransaction, (cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID, ))
		return Response(Success=True)

	def SET_SELL_TRIGGER(self, cmd: Command):

		user = self._database.GetUser(cmd.UserId)
		if user.error is not None:
			return self.error(cmd, "The user does not exist.")

		trig = self._database.GetTrigger(cmd.UserId, cmd.StockSymbol, "SELL")
		if trig.error is not None:
			return self.error(cmd, "User must set sell amount first.")

		trig.When = cmd.Amount
		self._database.AddNewTrigger(trig)

		return Response(Success=True)

	def CANCEL_SET_SELL(self, cmd: Command):
		user = self._database.GetUser(cmd.UserId)
		if user.error is not None:
			return self.error(cmd, "The user does not exist.")

		trig: Trigger = self._database.CancelTrigger(cmd.UserId, cmd.StockSymbol, "SELL")
		if trig.error is not None:
			return self.error(cmd, "No sell trigger to cancel")

		resp = self._database.UnreserveShares(cmd.UserId, cmd.StockSymbol, trig.Shares)
		if resp.error is not None:
			logging.error(resp.error)
			return self.error(cmd, "Internal error occured.")

		_executor.submit(self._audit.AccountTransaction, *(cmd.UserId, trig.Amount, "unreserve", cmd.TransactionID, ))
		return Response(Success=True)

	# TODO://
	def DUMPLOG(self, cmd: Command):
		return Response(Success=True)

	def DISPLAY_SUMMARY(self, cmd: Command):
		return Response(Success=True)


if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.DEBUG)
	
	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.DEBUG)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	trans = transactionserver(use_rpc=True, server=True)

if __name__ == "__main__":
	transactionserver(use_rpc=True, server=True)


