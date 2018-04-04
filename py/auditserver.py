import sys

sys.path.append('gen-py')


from Service import Service
from log import LoggerRpc


@Service(thrift_class=LoggerRpc, port=44422)
class AuditServer(object):
	def __init__(self, use_rpc=False, server=False):
		pass

	def UserCommand(self, cmd):
		return None

	def QuoteServer(self, quote, tid):
		return None

	def AccountTransaction(self, userid, funds, action, tid):
		return None

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
	audit = AuditServer(use_rpc=True, server=True)