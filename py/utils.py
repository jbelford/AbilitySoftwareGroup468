import logging
from concurrent.futures import ThreadPoolExecutor

from flask import json

_executor = ThreadPoolExecutor(max_workers=4)

def process_error(audit, cmd, msg):
	logging.debug(msg)
	_executor.submit(audit.ErrorEvent, args=(cmd, msg, ))
	return json.dumps({"Success": False, "Message": msg})


def thrift_to_json(obj):
	vars = obj.__dict__
	return {k:v for (k,v) in vars.items() if v is not None}