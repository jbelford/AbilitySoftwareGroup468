import logging
import sys


sys.path.append('gen-py')


from Service import Service
from log import LoggerRpc
from shared.ttypes import Command


@Service(thrift_class=LoggerRpc, port=44422)
class AuditServer(object):
	def __init__(self, use_rpc=False, server=False):
		pass

	def UserCommand(self, cmd: Command):
		logging.info(cmd)
		
	def QuoteServer(self, quote, tid):
		logging.info(quote, tid)

	def AccountTransaction(self, userid, funds, action, tid):
		logging.info(str(userid) + " " + str(funds)
		                    + " " + str(action) + " " + str(tid))
		
	def SystemEvent(self, cmd):
		pass

	def ErrorEvent(self, cmd):
		pass

	def DebugEvent(self, cmd):
		pass

	def DumpLog(self, cmd):
		return ""

	def DumpLogUser(self, cmd):
		return ""



if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.INFO)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.INFO)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	audit = AuditServer(use_rpc=True, server=True)