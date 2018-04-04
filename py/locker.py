import logging
import sys
from multiprocessing import Lock

import time

sys.path.append('gen-py')

from Service import Service
from lockerRPC import LockerRPC
from auditserver import AuditServer


@Service(thrift_class=LockerRPC, port=44421)
class Locker(object):

	def __init__(self, use_rpc=False, server=False, max_parallelism=1000):
		self._audit = AuditServer(use_rpc=use_rpc, server=False)
		self._quotes = [Lock()] * max_parallelism
		self._users = [Lock()] * max_parallelism
		self._triggers = [Lock()] * max_parallelism
		self._trans = [Lock()] * max_parallelism
		
		self._user_lock = Lock()
		self._quote_lock = Lock()
		self._trans_lock = Lock()
		self._trigger_lock = Lock()
		self._timeout = 1.0
		
	def isLocked(self, Key, Type):
		assert Type in ["USER", "QUOTE", "TRANSACTION", "TRIGGER"]
		
		is_locked = False
		
		return is_locked
	
	def requestLock(self, Key, Type):
		# Return unique key for that lock...
		if Type == "USER":
			self._user_lock.acquire(timeout=self._timeout/2)
			
			self._users[hash(Key) % len(self._users)].acquire(timeout=self._timeout)
			
			self._user_lock.release()
		elif Type == "QUOTE":
			self._quote_lock.acquire(timeout=self._timeout/2)
			
			self._quotes[hash(Key) % len(self._quotes)].acquire(timeout=self._timeout)
			
			self._quote_lock.release()
		elif Type == "TRANSACTION":
			self._trans_lock.acquire(timeout=self._timeout/2)
			
			self._trans[hash(Key) % len(self._trans)].acquire(timeout=self._timeout)
			
			self._trans_lock.release()
		elif Type == "TRIGGER":
			self._trigger_lock.acquire(timeout=self._timeout/2)
			
			self._triggers[hash(Key) % len(self._triggers)].acquire(timeout=self._timeout)
			
			self._trigger_lock.release()
			
		return hash(Key)
	
	def releaseLock(self, Key, Type):
		if Type == "USER":
			self._user_lock.acquire(timeout=self._timeout)
			
			self._users[hash(Key) % len(self._users)].release()
			
			self._user_lock.release()
		elif Type == "QUOTE":
			self._quote_lock.acquire(timeout=self._timeout)
			
			self._quotes[hash(Key) % len(self._quotes)].release()
			
			self._quote_lock.release()
		elif Type == "TRANSACTION":
			self._trans_lock.acquire(timeout=self._timeout)
			
			self._trans[hash(Key) % len(self._trans)].release()
			
			self._trans_lock.release()
		elif Type == "TRANSACTION":
			self._trigger_lock.acquire(timeout=self._timeout)
			
			self._triggers[hash(Key) % len(self._triggers)].release()
			
			self._trigger_lock.release()
			
if __name__ == "__main__":
	Locker(use_rpc=True, server=True)