import os
import pickle
import sys
from multiprocessing import Lock

from databaseRPC.ttypes import DBResponse
from shared.ttypes import Response, User, PendingTxn

sys.path.append('gen-py')

from Service import Service
from databaseRPC import Database


@Service(thrift_class=Database, port=44421)
class dbserver(object):
	tables_lookup = {
		"Users": 0,
		"Triggers": 1,
		"Transactions": 2,
	}

	tables = []
	lock = None
	_timeout = -1


	# TODO:// Make sure you don't return without unlocking!

	def __init__(self, use_rpc=False, server=False):
		self.tables = {}
		self.lock = Lock()
		self._timeout = 10  # default timeout of 10 seconds.

	def __init_tables(self):
		self.tables = [None] * len(self.tables_lookup.keys())

		for table, ind in self.tables_lookup:
			if os.path.exists(table + ".pkl"):
				with open(table + ".pkl") as my_pickle:
					self.tables[ind] = pickle.load(my_pickle)
			else:
				# Initialize to new dictionary.
				self.tables[ind] = {}

	def __save_table(self, table):
		with open(table + ".pkl") as my_pickle:
			pickle.dump(self.tables[self.tables_lookup[table]], my_pickle)

	def __get_key(self, table, key):
		assert table in self.tables_lookup.keys()
		my_table = self.tables[self.tables_lookup[table]]

		if key not in my_table.keys():
			# Does this update the original object?
			my_table.update({key: {}})
		return my_table[key]

	def __replace_key(self, table, key, value):
		assert table in self.tables_lookup.keys()
		my_table = self.tables[self.tables_lookup[table]]

		my_table.update({key: value})

		# TODO:// Mark as different, write it to a permanent log...

	def __get_new_user(self, userId):
		return {
			"userId": userId,
			"balance": 0,
			"reserved": 0,
			"stock": {},
			"triggers": []
		}

	def __get_new_trigger(self):
		return {}

	def __get_new_transaction(self):
		return {}

	def __lock_user(self, userId):
		self.lock.acquire(timeout=self._timeout)

	def __unlock_user(self, userId):
		self.lock.release()

	def __lock_txn(self, txn):
		self.lock.acquire(timeout=self._timeout)

	def __unlock_txn(self, txn):
		self.lock.release()

	def AddUserMoney(self, userId, amount):
		self.__lock_user(userId)

		user = self.__get_key("Users", userId)
		if user == {}:
			user = self.__get_new_user(userId)

		user["balance"] += amount
		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)

		# Error
		resp = DBResponse(
				user=User(User=userId, Balance=user["balance"], Reserved=0, stock=0)
		)

		return resp


	def UnreserveMoney(self, userId, amount):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)

		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")

		if user["reserved"] - amount < 0:
			self.__unlock_user(userId)
			return DBResponse(error="Not Enough Money To Unreserve!")

		user["reserved"] -= amount
		user["balance"] += amount

		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)
		return DBResponse()

	def ReserveMoney(self, userId, amount):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)

		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")

		if user["reserved"] - amount < 0:
			self.__unlock_user(userId)
			return DBResponse(error="Not Enough Money To Unreserve!")

		user["reserved"] -= amount
		user["balance"] += amount

		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)
		return DBResponse(user=User(User=userId,
		                            Balance=user["balance"],
		                            Reserved=user["reserved"]))

	def GetReserveMoney(self, userId):
		return self.GetUser(userId=userId).user.Reserved

	def GetReservedShares(self, userId, stock):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)

		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")
		if user["stocks"] == {} or \
				"shares." + stock + ".real" not in user["stock"].keys():
			self.__unlock_user(userId)
			return DBResponse(error="User does not own any of this stock.")

		st = user["stock"][stock]
		shares = st["shares." + stock + ".reserved"]

		self.__unlock_user(userId)
		return shares

	def UnreserveShares(self, userId, stock, shares):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)

		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")
		if user["stocks"] == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not own any stocks.")

		if "shares." + stock + ".real" not in user["stock"].keys():
			self.__unlock_user(userId)
			return DBResponse(error="User does not own stocks!")

		if user["stock"]["shares." + stock + ".real"] < shares:
			self.__unlock_user(userId)
			return DBResponse(error="User does not own that many shares")

		st = user["stock"][stock]
		st["shares." + stock + ".real"] += shares
		st["shares." + stock + ".reserved"] -= shares

		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)
		return DBResponse()

	def ReserveShares(self, userId, stock, shares):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)

		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")
		if user["stocks"] == {} or \
				"shares." + stock + ".real" not in user["stock"].keys() or \
				stock["shares." + stock + ".real"] == 0:
			self.__unlock_user(userId)
			return DBResponse(error="User does not own any of this stock.")
		else:
			st = user["stock"]
			if st["shares." + stock + ".real"] < shares:
				self.__unlock_user(userId)
				return DBResponse(error="User does not own enough of this stock.")
			st["shares." + stock + ".real"] -= shares
			st["shares." + stock + ".reserved"] += shares

		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)
		return DBResponse()

	def GetUser(self, userId):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)
		self.__unlock_user(userId)

		return DBResponse(user=User(User=userId,
		                            Balance=user["balance"],
		                            Reserved=user["reserved"]))

	def BulkTransactions(self, txns: [PendingTxn], wasCached):
		for txn in txns:
			if txn.Type == "Buy":
				self.buyParams(txn, wasCached)
			else:
				self.sellParams(txn, wasCached)

	def ProcessTxn(self, txn: PendingTxn, wasCached):
		if txn.Type == "Buy":
			self.buyParams(txn, wasCached)
		else:
			self.sellParams(txn, wasCached)

	def PushPendingTxn(self, pending: PendingTxn):
		key = pending.UserId + ":" + pending.Type
		self.__lock_txn(key)

		curr_pending = self.__get_key("Transactions", key)
		if curr_pending == {}:
			curr_pending = [pending]
		else:
			curr_pending.append(pending)
		self.__replace_key("Transactions", key, curr_pending)
		self.__unlock_txn(key)

	def PopPendingTxn(self, userId, txnType):
		key = userId + ":" + txnType
		self.__lock_txn(key)

		# TODO:// Ignore any that have expired...
		curr_pending = self.__get_key("Transactions", key)
		if curr_pending == {} or len(curr_pending) == 0:
			last: PendingTxn = None
		else:
			last: PendingTxn = curr_pending.pop()
		self.__replace_key("Transactions", key, curr_pending)
		self.__unlock_txn(key)

		return last


	# TODO:// ---------------------------
	def buyParams(self, txn, wasCached):
		self.__lock_txn(txn)

		if wasCached:
			pass
		else:
			pass

		self.__unlock_txn(txn)
		return None

	def sellParams(self, txn, wasCached):
		self.__lock_txn(txn)

		if wasCached:
			pass
		else:
			pass

		self.__unlock_txn(txn)
		return None

