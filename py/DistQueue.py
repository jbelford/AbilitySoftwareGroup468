import logging
import sys
from copy import deepcopy
from multiprocessing import Queue, Lock
from threading import Thread

import functools

sys.path.append('gen-py')

import time

from shared.ttypes import Command, Response
from Service import Service
from distqueue import DistQueueRPC


@Service(thrift_class=DistQueueRPC, port=44425)
class DistQueue(object):
	
	def __init__(self, use_rpc=False, server=False, is_master=False):
		self.num_queues = 1
		self.transaction_timeout = 10 * 100000  # timeout in (ms)
		self.transaction_queues = [Queue()] * self.num_queues
		self.current_incomplete = [[]] * self.num_queues
		self.completed = [{}] * self.num_queues
		self.timeouts = {}
		
		self.lock = Lock()
		
		self._is_master = is_master
		for q_i in range(self.num_queues):
			worker = Thread(target=functools.partial(self.resubmit_incomplete, q_i))
			worker.start()
	
	def __check_for_workers(self):
		pass
	
	def resubmit_incomplete(self, queue_to_check):
		# Thread that monitors queue, checks for timedout/failed commands, resubmit
		cur_queue = self.current_incomplete[queue_to_check]
		while True:
			to_remove = []
			
			if len(cur_queue) == 0:
				logging.debug("No work in progress, sleeping: " + str(queue_to_check))
				time.sleep(0.5)
			
			for job in cur_queue:
				cur_job: Command = job
				if cur_job.TransactionID in self.timeouts.keys():
					if self.timeouts[cur_job.TransactionID] - time.time() >= self.transaction_timeout:
						to_remove.append(cur_job)
						logging.warning(str(cur_job.TransactionID) + " timed out for errored.")
			
			for r in to_remove:
				job: Command = r
				if job.TransactionID in self.timeouts.keys():
					self.current_incomplete[queue_to_check].remove(job)
					self.timeouts.__delitem__(job.TransactionID)
					# Resubmit...
					logging.info("Resubmitting Job: " + str(job.TransactionID))
					self.PutItem(queue_to_check, job)
	
	def GetCompletedItem(self, QueueInst, cmd_id: int):
		if cmd_id in self.completed[QueueInst % self.num_queues].keys():
			res = self.completed[QueueInst % self.num_queues][cmd_id]
			self.completed[QueueInst % self.num_queues][cmd_id].remove(cmd_id)
			return res
		else:
			return Response(Success=False, Message="No Response Found.")
	
	def GetItem(self, QueueInst):
		self.lock.acquire(timeout=5)
		item: Command = self.transaction_queues[QueueInst % self.num_queues].get()
		self.current_incomplete[QueueInst % self.num_queues].append(item)
		self.lock.release()
		print(item)
		if item is None:
			return Command(C_type=-1)
		# Mark the 'checkout' time.
		self.timeouts.update({item.TransactionID: time.time()})
		print(item)
		return item
	
	def MarkComplete(self, QueueInst, cmd: Command, res: Response):
		if cmd.TransactionID in self.timeouts.keys():
			self.timeouts.pop(cmd.TransactionID)
			self.current_incomplete[QueueInst % self.num_queues].remove(cmd)
			self.completed[QueueInst % self.num_queues].update({cmd.TransactionID: res})
	
	def PutItem(self, QueueInst, cmd: Command):
		#self.lock.acquire()
		self.transaction_queues[QueueInst % self.num_queues].put(cmd)
		#self.lock.release()
		return Response(Success=True, Message=str(cmd.TransactionID) + " in Progress...")
		
if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.DEBUG)
	
	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.DEBUG)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)
	
	DistQueue(use_rpc=True, server=True, is_master=True)
