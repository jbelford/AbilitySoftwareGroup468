import logging
from Service import Service


class Seek(Service):
	def __init__(self, thrift_class, port):
		super().__init__(thrift_class, port)
	
	def __call__(self, original_clazz):
		logging.debug("Wrapped " + str(original_clazz.__name__))
		
		decorator_self = self
		
		def wrappee(*args, **kwargs):
			logging.debug('in decorator before wrapee with flag ' + str(decorator_self._thrift_class.__name__))
			
			dec_name = str(decorator_self._thrift_class.__name__).split(".")[-1]
			print(dec_name)
			
			assert "use_rpc" in kwargs.keys() and "server" in kwargs.keys(), \
				"Wrapped function must be subclass of RPC."
			
			if "use_rpc" in kwargs.keys() and "server" in kwargs.keys() \
					and kwargs["use_rpc"] and kwargs["server"]:
				
				from thrift.protocol import TBinaryProtocol
				from thrift.transport import TSocket
				from thrift.transport import TTransport
				from thrift.server import TServer
				
				handler = original_clazz(*args, **kwargs)
				processor = decorator_self._thrift_class.Processor(handler)
				transport = TSocket.TServerSocket(port=decorator_self._port)
				tfactory = TTransport.TBufferedTransportFactory()
				pfactory = TBinaryProtocol.TBinaryProtocolFactory()
				
				self.__inst = handler
				
				server = TServer.TThreadPoolServer(processor, transport,
				                                   tfactory, pfactory)
				server.setNumThreads(cpu_count() * 4)
				
				logging.info("Serving: " + dec_name)
				server.serve()
				logging.info('Done: ' + dec_name)
			
			elif "use_rpc" in kwargs.keys() and "server" in kwargs.keys() \
					and kwargs["use_rpc"] and not kwargs["server"]:
				
				from thrift.protocol import TBinaryProtocol
				from thrift.transport import TSocket
				from thrift.transport import TTransport
				from thrift.server import TServer
				
				# Make socket
				self._transport = TSocket.TSocket(host=dec_name,
				                                  port=decorator_self._port)
				# Buffering is critical. Raw sockets are very slow
				self._transport = TTransport.TBufferedTransport(self._transport)
				# Connect!
				self._transport.open()
				# Wrap in a protocol
				protocol = TBinaryProtocol.TBinaryProtocol(self._transport)
				# Create a client to use the protocol encoder
				client = decorator_self._thrift_class.Client(protocol)
				logging.info("Client (" + dec_name + ") connected to server: " + str(
					self._transport.isOpen()))
				
				return client
			
			else:
				return original_clazz(*args, **kwargs)
		
		logging.debug('in decorator after wrapee with flag ' + str(decorator_self._thrift_class.__name__))
		return wrappee
	
	def __del__(self):
		if self._transport is not None:
			self._transport.close()
