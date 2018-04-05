import logging
import os
import pickle
import sys
from copy import deepcopy
from threading import Thread

import time

sys.path.append('gen-py')

from multiprocessing import Lock, Queue

from databaseRPC.ttypes import DBResponse
from shared.ttypes import User, PendingTxn, Trigger
from Service import Service
from locker import Locker
from databaseRPC import Database


@Service(thrift_class=Database, port=44423)
class dbserver(object):
	_tables_lookup = {
		"Users": 0,
		"Triggers": 1,
		"Transactions": 2,
	}

	_tables = []
	_timeout = -1


	# TODO:// Make sure you don't return without unlocking!

	def __init__(self, use_rpc=False, server=False):
		self._tables = {}
		self._timeout = 10000  # default timeout of 10 seconds.
		self.__init_tables()
		self._lock = Locker(use_rpc=use_rpc, server=False)
		self._my_lock = Lock()
		
		self._update_queue = Queue()
		
		
		self.__num_tables_to_keep = 10
		t = Thread(target=self.__poll_for_table_changes, args=(self._update_queue, ))
		t.start()
	
	
	def __init_tables(self):
		self._tables = [None] * len(self._tables_lookup.keys())
		
		logging.debug("Initializing Tables.")
		# TODO:// Read in split _tables properly
		for table, ind in self._tables_lookup.items():
			if os.path.exists(table + ".pkl"):
				with open(table + ".pkl", "r") as my_pickle:
					self._tables[ind] = pickle.load(my_pickle)
			else:
				# Initialize to new dictionary.
				self._tables[ind] = {}
	
	def __poll_for_table_changes(self, queue):
		while True:
			table_to_update = queue.get()
			logging.debug("Saving Table: " + str(table_to_update))
			self.__save_table(table_to_update)
			
			
	def __save_table(self, table):
		logging.debug("Save Table")
		
		# Save the partition of the table to file...
		with open(table + ".pkl", "wb") as my_pickle:
			table_name = table.split("_")[0]
			my_table = {k:v for (k, v) in self._tables[self._tables_lookup[table_name]].items() if
			            str(table_name) + "_" + str(hash(k) % self.__num_tables_to_keep) == table}
			print(my_table)
			
			pickle.dump(my_table, my_pickle)

	def __get_key(self, table, key):
		assert table in self._tables_lookup.keys()
		my_table = self._tables[self._tables_lookup[table]]

		if key not in my_table.keys():
			# Does this update the original object?
			my_table.update({key: {}})
		return my_table[key]

	def __replace_key(self, table, key, value):
		assert table in self._tables_lookup.keys()
		my_table = self._tables[self._tables_lookup[table]]

		my_table.update({key: value})

		# TODO:// Mark as different, write it to a permanent log...
		self._update_queue.put(str(table) + "_" + str(hash(key) % self.__num_tables_to_keep))
		
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
		#self._my_lock.acquire()
		#self._lock.requestLock(userId, "USER")
		#self._my_lock.release()
		pass

	def __unlock_user(self, userId):
		#self._my_lock.acquire()
		#self._lock.releaseLock(userId, "USER")
		#self._my_lock.release()
		pass
		
	def __lock_trigger(self, txn):
		#self._my_lock.acquire()
		#self._lock.requestLock(txn, "TRIGGER")
		#self._my_lock.release()
		pass

	def __unlock_triggers(self, txn):
		#self._my_lock.acquire()
		#self._lock.releaseLock(txn, "TRIGGER")
		#self._my_lock.release()
		pass

	def __lock_txn(self, txn):
		#self._lock.requestLock(txn, "TRANSACTION")
		pass

	def __unlock_txn(self, txn):
		#self._lock.releaseLock(txn, "TRANSACTION")
		pass

	def AddUserMoney(self, userId, amount):
		self.__lock_user(userId)

		user = self.__get_key("Users", userId)
		if user == {}:
			logging.debug("Getting new user.")
			user = self.__get_new_user(userId)
		
		logging.debug("Adding balance")
		new_balance = user["balance"] + amount
		logging.debug("Updating balance")
		user["balance"] = new_balance
		
		logging.debug("Replacing Key")
		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)

		# Error
		logging.debug("Making User")

		my_user = User(User=userId, Balance=new_balance)
		
		logging.debug("Making Response")
		resp = DBResponse(user=my_user)
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
		
		new_reserved = user["reserved"] - amount
		new_balance = user["balance"] + amount
		user["reserved"] = new_reserved
		user["balance"] = new_balance

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
		new_reserved = user["reserved"] - amount
		new_balance = user["balance"] + amount
		
		user["reserved"] = new_reserved
		user["balance"] = new_balance

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
			return -1
		if "stocks" not in user.keys() or user["stocks"] == {} or \
				"shares." + stock + ".real" not in user["stock"].keys():
			self.__unlock_user(userId)
			return -1

		st = user["stock"][stock]
		shares = st["shares." + stock + ".reserved"]

		self.__unlock_user(userId)
		return shares

	def UnreserveShares(self, userId, stock, shares):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)
		
		real_key = "shares." + stock + ".real"
		reserved_key = "shares." + stock + ".reserved"
		
		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")
		if user["stocks"] == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not own any stocks.")

		if real_key not in user["stock"].keys():
			self.__unlock_user(userId)
			return DBResponse(error="User does not own stocks!")
		if user["stock"][real_key] < shares:
			self.__unlock_user(userId)
			return DBResponse(error="User does not own that many shares")

		st = user["stock"][stock]
		
		real_shares = st[real_key] + shares
		reserved_shares = st[reserved_key] - shares
		st[real_key] = real_shares
		st[reserved_key] = reserved_shares

		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)
		return DBResponse()

	def ReserveShares(self, userId, stock, shares):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)
		
		real_key = "shares." + stock + ".real"
		reserved_key = "shares." + stock + ".reserved"
		
		if user == {}:
			self.__unlock_user(userId)
			return DBResponse(error="User does not exist.")
		if user["stocks"] == {} or \
				real_key not in user["stock"].keys() or \
				stock[real_key] == 0:
			self.__unlock_user(userId)
			return DBResponse(error="User does not own any of this stock.")
		else:
			st = user["stock"]
			if st[real_key] < shares:
				self.__unlock_user(userId)
				return DBResponse(error="User does not own enough of this stock.")
			
			real_shares = st[real_key] - shares
			reserved_shares = st[reserved_key] + shares
			
			st[real_key] = real_shares
			st[reserved_key] = reserved_shares

		self.__replace_key("Users", userId, user)

		self.__unlock_user(userId)
		return DBResponse()

	def GetUser(self, userId):
		self.__lock_user(userId)
		user = self.__get_key("Users", userId)
		self.__unlock_user(userId)
		
		if user == {}:
			return DBResponse(error="The user does not exist.")
		
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
			
		return DBResponse()

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
			last: PendingTxn = PendingTxn(error="No Transaction Found.")
		else:
			last: PendingTxn = curr_pending.pop()
		self.__replace_key("Transactions", key, curr_pending)
		self.__unlock_txn(key)

		return last

	def AddNewTrigger(self, trigger: Trigger):
		self.__lock_trigger(trigger.UserId)

		key = trigger.UserId + ":" + trigger.Stock + ":" + trigger.Type

		self.__replace_key("Triggers", key, Trigger.__dict__)

		self.__unlock_triggers(trigger.UserId)
		return DBResponse()

	def CancelTrigger(self, userId, stock, trigger_type):
		key = userId + ":" + stock + ":" + trigger_type
		self.__lock_trigger(key)

		trig = deepcopy(self.__get_key("Triggers", key))
		if trig is None:
			self.__unlock_triggers(key)
			return Trigger(error="Trigger does not exist.")
		self.__replace_key("Triggers", key, None)  # Write a none there to cancel it...
		self.__unlock_triggers(key)
		
		if trig == {}:
			return Trigger(error="Trigger does not exist.")
		
		return trig

	def GetTrigger(self, userId, stock, trigger_type):
		key = userId + ":" + stock + ":" + trigger_type
		self.__lock_trigger(key)

		trig = self.__get_key("Triggers", key)
		
		self.__unlock_triggers(key)
		if trig == {}:
			return Trigger(error="Trigger does not exist.")
		return trig


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

if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.INFO)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.INFO)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	dbserver(use_rpc=True, server=True)
