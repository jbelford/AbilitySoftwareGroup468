import logging
import sys
sys.path.append('gen-py')

from shared.ttypes import PendingTxn


from Service import Service
from auditserver import AuditServer
from cache import Cache
from database import dbserver
from transactionRPC import TriggerManRpc


@Service(thrift_class=TriggerManRpc, port=44424)
class TriggerMan(object):

	# TODO:// Lock on the user...

	def __init__(self, use_rpc=False, server=False):
		self._database = dbserver(use_rpc=use_rpc, server=False)
		self._audit = AuditServer(use_rpc=use_rpc, server=False)
		self._cache = Cache(use_rpc=use_rpc, server=False, mock=True)

		# Start the separate thread...

	def _start(self):
		pass

	def ProcessTrigger(self, t):
		return PendingTxn()


if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.NOTSET)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.NOTSET)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	TriggerMan(use_rpc=True, server=True)
