import sys
sys.path.append('gen-py')

from multiprocessing import Queue
import time

from shared.ttypes import Command
from Service import Service
from distqueue import DistQueueRPC


@Service(thrift_class=DistQueueRPC, port=44425)
class DistQueue(object):
	
	def __init__(self, use_rpc=False, server=False, is_master=False):
		self.num_queues = 10
		self.transaction_timeout = 10  # timeout in (ms)
		self.transaction_queues = [Queue()] * self.num_queues
		self.current_incomplete = [Queue()] * self.num_queues
		self.timeouts = {}
		
		self._is_master = is_master
	
	
	
	def __check_for_workers(self):
		pass
	
	def __resubmit_incomplete(self):
		# Thread that monitors queue, checks for timedout/failed commands, resubmit
		pass
	
	def GetItem(self, QueueInst):
		item:Command = self.transaction_queues[QueueInst].get()
		# Mark the 'checkout' time.
		self.timeouts.update({item.TransactionID: time.time()})
		return item
	
	def MarkComplete(self, QueueInst, cmd: Command):
		self.timeouts.__delitem__(cmd.TransactionID)
		self.current_incomplete[QueueInst].remove(cmd)
	
	def PutItem(self, QueueInst, cmd: Command):
		self.current_incomplete[QueueInst].put(cmd)
	
if __name__ == "__main__":
	DistQueue(use_rpc=True, server=True, is_master=True)
