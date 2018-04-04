import logging
import sys
sys.path.append('gen-py')


from Service import Service
from log import LoggerRpc


@Service(thrift_class=LoggerRpc, port=44422)
class AuditServer(object):
	def __init__(self, use_rpc=False, server=False):
		pass

	def UserCommand(self, cmd):
		return logging.info(cmd)

	def QuoteServer(self, quote, tid):
		return logging.info(quote, tid)

	def AccountTransaction(self, userid, funds, action, tid):
		return logging.info(userid, funds, action, tid)

	def SystemEvent(self, cmd):
		return None

	def ErrorEvent(self, cmd):
		return None

	def DebugEvent(self, cmd):
		return None

	def DumpLog(self, cmd):
		return None

	def DumpLogUser(self, cmd):
		return None



if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.NOTSET)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.NOTSET)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	audit = AuditServer(use_rpc=True, server=True)