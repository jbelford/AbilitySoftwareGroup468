import logging
import sys
sys.path.append('gen-py')

import time
from concurrent.futures import ThreadPoolExecutor
from datetime import timedelta
from functools import update_wrapper

from flask import current_app, request, make_response, json, Flask

from CmdType import Cmd
from auditserver import AuditServer
from shared.ttypes import Command
from transaction import transactionserver
from utils import process_error, thrift_to_json


def crossdomain(origin=None, methods=None, headers=None, max_age=21600,
				attach_to_all=True, automatic_options=True):
	"""Decorator function that allows crossdomain requests.
	  Courtesy of
	  https://blog.skyred.fi/articles/better-crossdomain-snippet-for-flask.html
	"""
	if methods is not None:
		methods = ', '.join(sorted(x.upper() for x in methods))
	if isinstance(max_age, timedelta):
		max_age = max_age.total_seconds()

	def get_methods():
		""" Determines which methods are allowed
		"""
		if methods is not None:
			return methods

		options_resp = current_app.make_default_options_response()
		return options_resp.headers['allow']

	def decorator(f):
		"""The decorator function
		"""
		def wrapped_function(*args, **kwargs):
			"""Caries out the actual cross domain code
			"""
			if automatic_options and request.method == 'OPTIONS':
				resp = current_app.make_default_options_response()
			else:
				resp = make_response(f(*args, **kwargs))
			if not attach_to_all and request.method != 'OPTIONS':
				return resp

			h = resp.headers
			h['Access-Control-Allow-Origin'] = origin
			h['Access-Control-Allow-Methods'] = get_methods()
			h['Access-Control-Max-Age'] = str(max_age)
			h['Access-Control-Allow-Credentials'] = 'true'
			h['Access-Control-Allow-Headers'] = \
				"Origin, X-Requested-With, Content-Type, Accept, Authorization"
			if headers is not None:
				h['Access-Control-Allow-Headers'] = headers
			return resp

		f.provide_automatic_options = False
		return update_wrapper(wrapped_function, f)
	return decorator


txn = transactionserver(use_rpc=True, server=False)
audit = AuditServer(use_rpc=True, server=False)
executor = ThreadPoolExecutor(max_workers=4)

app = Flask(__name__)
port = 44420
hostname = "0.0.0.0"

@app.route('/<t_id>/<user_id>/display_summary', methods=["GET"])
@crossdomain(origin='*')
def display_summary(t_id, user_id):
	# TODO://
	logging.warning("NOT IMPLEMENTED")
	return json.dumps({})


@app.route('/<t_id>/<user_id>/add', methods=["POST"])
@crossdomain(origin='*')
def add(t_id, user_id):
	cmd = Command(TransactionID=t_id, C_type=Cmd.ADD.value, UserId=user_id, Timestamp=time.time())
	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")
	
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.ADD.value,
	              UserId=user_id, Timestamp=time.time(), Amount=amount)
	
	log = executor.submit(audit.UserCommand, (cmd, ))
	resp = txn.ADD(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/quote', methods=["GET"])
@crossdomain(origin='*')
def quote(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.QUOTE.value, UserId=user_id, Timestamp=time.time(), StockSymbol=stock)
	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	log = executor.submit(audit.UserCommand, (cmd, ))
	resp = txn.QUOTE(cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/buy', methods=["POST"])
@crossdomain(origin='*')
def buy(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.BUY.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)

	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")

	cmd = Command(TransactionID=int(t_id), C_type=Cmd.BUY.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.BUY(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/commit_buy', methods=["POST"])
@crossdomain(origin='*')
def commit_buy(t_id, user_id):
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.COMMIT_BUY.value, UserId=user_id,
	              Timestamp=time.time())

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.COMMIT_BUY(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/cancel_buy', methods=["POST"])
@crossdomain(origin='*')
def cancel_buy(t_id, user_id):
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.CANCEL_BUY.value, UserId=user_id,
	              Timestamp=time.time())

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.CANCEL_BUY(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/sell', methods=["POST"])
@crossdomain(origin='*')
def sell(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SELL.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)
	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(cmd, "Parameter: 'amount' must be greater than 0")
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SELL.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.SELL(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/commit_sell', methods=["POST"])
@crossdomain(origin='*')
def commit_sell(t_id, user_id):
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.COMMIT_SELL.value, UserId=user_id,
	              Timestamp=time.time())

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.COMMIT_SELL(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/cancel_sell', methods=["POST"])
@crossdomain(origin='*')
def cancel_sell(t_id, user_id):
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.CANCEL_SELL.value, UserId=user_id,
	              Timestamp=time.time())

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.CANCEL_SELL(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/set_buy_amount', methods=["POST"])
@crossdomain(origin='*')
def set_buy_amount(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_BUY_AMOUNT.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)
	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_BUY_AMOUNT.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.SET_BUY_AMOUNT(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/cancel_set_buy', methods=["POST"])
@crossdomain(origin='*')
def cancel_set_buy(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.CANCEL_SET_BUY.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)
	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.CANCEL_SET_BUY(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/set_buy_trigger', methods=["POST"])
@crossdomain(origin='*')
def set_buy_trigger(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_BUY_TRIGGER.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)
	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_BUY_TRIGGER.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.SET_BUY_TRIGGER(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/set_sell_amount', methods=["POST"])
@crossdomain(origin='*')
def set_sell_amount(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_SELL_AMOUNT.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)

	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")

	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_SELL_AMOUNT.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)

	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.SET_SELL_AMOUNT(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/set_sell_trigger', methods=["POST"])
@crossdomain(origin='*')
def set_sell_trigger(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_SELL_TRIGGER.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)

	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")
	
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_BUY_TRIGGER.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)
	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.SET_SELL_TRIGGER(cmd=cmd)
	log.result()
	if resp is None:
		return process_error(audit, cmd, "Internal Error prevented operation")
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/cancel_set_sell', methods=["POST"])
@crossdomain(origin='*')
def cancel_set_sell(t_id, user_id):
	stock = request.args.get('stock', default = '', type = str)
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.CANCEL_SET_SELL.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock)

	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")

	try:
		amount = int(request.args.get("amount", default=0, type=int))
	except:
		return process_error(audit, cmd, "Count not process field: 'amount'")
	if amount <= 0:
		return process_error(audit, cmd, "Parameter: 'amount' must be greater than 0")
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.SET_BUY_TRIGGER.value, UserId=user_id,
	              Timestamp=time.time(), StockSymbol=stock, Amount=amount)
	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.CANCEL_SET_SELL(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


@app.route('/<t_id>/<user_id>/dumplog', methods=["GET"])
@crossdomain(origin='*')
def dumplog(t_id, user_id):
	filename = request.form["filename"]
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.DUMPLOG.value, UserId=user_id,
	              Timestamp=time.time(), FileName=filename)

	if filename == "" and filename != "admin":
		return process_error(audit, cmd, "Parameter: 'filename' cannot be empty.")


	log = executor.submit(audit.UserCommand, (cmd,))
	resp = txn.DUMPLOG(cmd=cmd)
	log.result()
	return json.dumps(thrift_to_json(resp))


def start_web_server():
	app.run(host=hostname, port=port)

if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.NOTSET)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.NOTSET)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	start_web_server()
