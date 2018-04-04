import logging
import sys
sys.path.append('gen-py')

import time
from multiprocessing import Lock

from shared.ttypes import QuoteData


from utils import process_error, _executor
from auditserver import AuditServer
from Service import Service
from locker import Locker
from cacheRPC import Cache
import socket


@Service(thrift_class=Cache, port=44425)
class Cache(object):
	quotes = {}

	def __init__(self, mock=True, timeout=10000, use_rpc=False, server=False):
		self.quotes = {}
		if not mock:
			self._quote_server_conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
			self._quote_server_conn.connect(('quoteserve.seng', 4444))
		else:
			self._quote_server_conn = None
		self._mock = mock
		self._audit = AuditServer(use_rpc=use_rpc, server=False)
		self._lock = Locker(use_rpc=use_rpc, server=False)
		self._timeout = timeout  # default timeout of 10 seconds.

	def error(self, cmd, msg):
		return process_error(self._audit, cmd, msg)

	def __del__(self):
		self._quote_server_conn.close()

	def __get_quote(self, quote, userId):
		if quote + "." + userId in self.quotes.keys():
			# Wait a full minute, then ignore it...
			if self.quotes[quote + "." + userId][0] + 60 <= time.time():
				return self.quotes[quote + "." + userId][1]
		return None

	def __set_quote(self, quote, userId, value, expiry):
		self.quotes.update({quote + "." + userId: [expiry, value]})


	# TODO:// Have multiple locks...
	def __lock_quote(self, quote, userId):
		#self._lock.requestLock(quote + ":" + userId, "QUOTE")
		pass

	def __unlock_quote(self, quote, userId):
		#self._lock.releaseLock(quote + ":" + userId, "QUOTE")
		pass

	def GetQuote(self, symbol, userId, tid):
		logging.debug("Getting Quote: " + str(symbol))
		# Get a line of text from the user
		fromUser = str(symbol) + ", " + str(userId) + "\n"
		if not self._mock:
			# Connect the socket
			# Send the user's query
			self._quote_server_conn.send(fromUser)
			# Read and print up to 1k of data.
			data = self._quote_server_conn.recv(1024)
			quote, sym, userid, timestamp, cryptokey = data.split(",")
		else:
			quote, sym, userid, timestamp, cryptokey = (12.55, symbol, userId, time.time(), "<<<MY_CRYPTOKEY>>>")
			_executor.submit(self._audit.QuoteServer, args=(quote, tid,))

		return QuoteData(UserId=userid, Symbol=symbol, Quote=quote, Timestamp=timestamp, Cryptokey=cryptokey)

if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.NOTSET)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.NOTSET)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	Cache(use_rpc=True, server=True)
