import logging
import sys
sys.path.append('gen-py')

import time

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

	# TODO:// Lock on the user...

	def __init__(self, use_rpc=False, server=False):
		self._database = dbserver(use_rpc=use_rpc, server=False)
		self._audit = AuditServer(use_rpc=use_rpc, server=False)
		self._cache = Cache(use_rpc=use_rpc, server=False, mock=True)

	def error(self, cmd, msg):
		return process_error(self._audit, cmd, msg)

	def ADD(self, cmd: Command):
		resp = self._database.AddUserMoney(userId=cmd.UserId, amount=cmd.Amount)
		err = resp.error
		if err is not None:
			return self.error(cmd, "Failed to create and/or add money to account")

		_executor.submit(self._audit.AccountTransaction, args=(cmd.UserId, cmd.Amount, "add", cmd.TransactionID))

		return Response(Success=True)

	def QUOTE(self, cmd: Command):
		resp: QuoteData = self._cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)

		if resp.error is not None:
			return self.error(cmd, "Quote server failed to respond with quote")

		return Response(Success=True, Quote=resp.Quote, Stock=resp.Symbol)

	def BUY(self, cmd: Command):
		resp: DBResponse = self._database.GetUser(userId=cmd.UserId)
		if resp.error is None:
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

		log = _executor.submit(self._audit.AccountTransaction, args=(cmd.UserId, cmd.Amount, "remove", cmd.TransactionID, ))
		# TODO:// Log to database.
		log.result()

		return Response(Success=True, Stock=buy.Stock, Shares=buy.Shares, Paid=buy.Price)

	def CANCEL_BUY(self, cmd):
		buy = self._database.PopPendingTxn(cmd.UserId, "BUY")
		if buy is None:
			return self.error(cmd, "There is no buy to cancel.")

		return Response(Success=True, Stock=buy.Stock, Shares=buy.Shares)

	def SELL(self, cmd: Command):
		resp = self._database.GetUser(cmd.UserId)
		if resp.error is not None:
			return self.error(cmd, "The user " + str(cmd.UserId) + " does not exist.")
		elif resp.user.stock[cmd.StockSymbol]["real"] == 0:
			return self.error(cmd, "User does not own any shares for that stock")

		quote: QuoteData = self._cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
		if quote.error is not None:
			return self.error(cmd, "Failed to get quote for that stock.")

		actual_shares = cmd.Amount // quote.Quote
		shares = actual_shares
		if shares <= 0:
			return self.error(cmd, "A single share is worth more than specified amount.")
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
		sell = self._database.PopPendingTxn(cmd.UserId, "SELL")
		if sell is None:
			return self.error(cmd, "There are no pending transactions.")

		resp:DBResponse = self._database.ProcessTxn(sell, True)
		if resp.error is not None:
			return self.error(cmd, "User no longer has the correct number of shares.")

		_executor.submit(self._audit.AccountTransaction, args=(cmd.UserId, cmd.Amount, "add", cmd.TransactionID))

		# TODO:// Log into database?... (trying to just write to file, see how it works)

		return Response(Success=True, Stock=sell.Stock, Shares=sell.Shares, Received=sell.Price)

	def CANCEL_SELL(self, cmd: Command):
		sell = self._database.PopPendingTxn(cmd.UserId, "SELL")
		if sell is None:
			return self.error(cmd, "There is no sell to cancel")

		return Response(Success=True, Stock=sell.Stock, Shares=sell.Shares)

	def SET_BUY_AMOUNT(self, cmd: Command):
		resp = self._database.GetUser(cmd.UserId)
		if resp.error is not None:
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

		_executor.submit(self._audit.AccountTransaction, args=(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID, ))
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

		_executor.submit(self._audit.AccountTransaction, args=(cmd.Amount, trig.Amount, "unreserve", cmd.TransactionID, ))

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
		_executor.submit(self._audit.AccountTransaction, args=(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID, ))
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

		_executor.submit(self._audit.AccountTransaction, args=(cmd.UserId, trig.Amount, "unreserve", cmd.TransactionID,))
		return Response(Success=True)

	# TODO://
	def DUMPLOG(self, cmd: Command):
		return Response(Success=True)

	def DISPLAY_SUMMARY(self, cmd: Command):
		return Response(Success=True)


if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.NOTSET)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.NOTSET)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	trans = transactionserver(use_rpc=True, server=True)
