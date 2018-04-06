import logging
import sys
sys.path.append('gen-py')

import time
from concurrent.futures import ThreadPoolExecutor
from datetime import timedelta
from functools import update_wrapper

from flask import current_app, request, make_response, json, Flask


from DistQueue import DistQueue
from CmdType import Cmd
from auditserver import AuditServer
from shared.ttypes import Command, Response
from transaction import transactionserver
from utils import process_error, thrift_to_json

use_rpc = True
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


txn = transactionserver(use_rpc=use_rpc, server=False)
audit = AuditServer(use_rpc=use_rpc, server=False)
executor = ThreadPoolExecutor(max_workers=4)
queue = DistQueue(use_rpc=use_rpc, server=False)
app = Flask(__name__)
port = 44420
hostname = "0.0.0.0"


def check_stock_and_amount(t_id, user_id, request, request_type):
	amount = request.args.get('amount', type=int)
	cmd0 = Command(TransactionID=int(t_id), C_type=request_type, UserId=user_id, Timestamp=time.time())
	if amount == "":
		return process_error(audit, cmd0, "Parameter: 'amount' cannot be empty.")
	
	stock = request.args.get('stock', default='', type=str)
	if stock == "":
		return process_error(audit, cmd0, "Parameter: 'stock' cannot be empty.")
	
	cmd = Command(TransactionID=int(t_id), C_type=request_type, UserId=user_id, Timestamp=time.time(), StockSymbol=stock, Amount=amount)
	return send_cmd(cmd)


def check_amount(t_id, user_id, request, request_type):
	amount = request.args.get('amount', type=int)
	cmd0 = Command(TransactionID=int(t_id), C_type=request_type, UserId=user_id, Timestamp=time.time())
	if amount == "":
		return process_error(audit, cmd0, "Parameter: 'amount' cannot be empty.")
	cmd = Command(TransactionID=int(t_id), C_type=request_type, UserId=user_id, Timestamp=time.time(), Amount=amount)
	return send_cmd(cmd)
	

def check_stock(t_id, user_id, request, request_type):
	stock = request.args.get('stock', default='', type=str)
	cmd = Command(TransactionID=int(t_id), C_type=request_type, UserId=user_id, Timestamp=time.time(), StockSymbol=stock)
	if stock == "":
		return process_error(audit, cmd, "Parameter: 'stock' cannot be empty.")
	return send_cmd(cmd)


def check_nothing(t_id, user_id, request_type):
	cmd = Command(TransactionID=int(t_id), C_type=request_type, UserId=user_id, Timestamp=time.time())
	return send_cmd(cmd)


def send_cmd(cmd):
	#log = executor.submit(audit.UserCommand, *(cmd,))
	logging.debug("Sending Transaction: " + str(cmd.TransactionID))
	resp = queue.PutItem(hash(cmd.UserId), cmd)
	logging.debug("Added to queue: " + str(cmd.TransactionID))
	return json.dumps(thrift_to_json(resp)) if type(resp) is Response else resp

# log.result()

@app.route('/<t_id>/<user_id>/get_result', methods=["GET"])
@crossdomain(origin="*")
def get_completed_result(t_id, user_id):
	resp = queue.GetCompletedItem(hash(user_id), t_id)
	return json.dumps(thrift_to_json(resp)) if type(resp) is Response else resp

@app.route('/<t_id>/<user_id>/display_summary', methods=["GET"])
@crossdomain(origin='*')
def display_summary(t_id, user_id):
	# TODO://
	logging.warning("NOT IMPLEMENTED")
	return json.dumps({})


@app.route('/<t_id>/<user_id>/add', methods=["POST"])
@crossdomain(origin='*')
def add(t_id, user_id):
	return check_amount(t_id, user_id, request, Cmd.ADD.value)


@app.route('/<t_id>/<user_id>/quote', methods=["GET"])
@crossdomain(origin='*')
def quote(t_id, user_id):
	return check_stock(t_id, user_id, request, Cmd.QUOTE.value)


@app.route('/<t_id>/<user_id>/buy', methods=["POST"])
@crossdomain(origin='*')
def buy(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.BUY.value)


@app.route('/<t_id>/<user_id>/commit_buy', methods=["POST"])
@crossdomain(origin='*')
def commit_buy(t_id, user_id):
	return check_nothing(t_id, user_id, Cmd.COMMIT_BUY.value)


@app.route('/<t_id>/<user_id>/cancel_buy', methods=["POST"])
@crossdomain(origin='*')
def cancel_buy(t_id, user_id):
	return check_nothing(t_id, user_id, Cmd.CANCEL_BUY.value)


@app.route('/<t_id>/<user_id>/sell', methods=["POST"])
@crossdomain(origin='*')
def sell(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.SELL.value)


@app.route('/<t_id>/<user_id>/commit_sell', methods=["POST"])
@crossdomain(origin='*')
def commit_sell(t_id, user_id):
	return check_nothing(t_id, user_id, Cmd.COMMIT_SELL.value)


@app.route('/<t_id>/<user_id>/cancel_sell', methods=["POST"])
@crossdomain(origin='*')
def cancel_sell(t_id, user_id):
	return check_nothing(t_id, user_id, Cmd.CANCEL_SELL.value)


@app.route('/<t_id>/<user_id>/set_buy_amount', methods=["POST"])
@crossdomain(origin='*')
def set_buy_amount(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.SET_BUY_AMOUNT.value)


@app.route('/<t_id>/<user_id>/cancel_set_buy', methods=["POST"])
@crossdomain(origin='*')
def cancel_set_buy(t_id, user_id):
	return check_stock(t_id, user_id, request, Cmd.CANCEL_SET_BUY.value)


@app.route('/<t_id>/<user_id>/set_buy_trigger', methods=["POST"])
@crossdomain(origin='*')
def set_buy_trigger(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.SET_BUY_TRIGGER.value)


@app.route('/<t_id>/<user_id>/set_sell_amount', methods=["POST"])
@crossdomain(origin='*')
def set_sell_amount(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.SET_SELL_AMOUNT.value)


@app.route('/<t_id>/<user_id>/set_sell_trigger', methods=["POST"])
@crossdomain(origin='*')
def set_sell_trigger(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.SET_SELL_TRIGGER.value)


@app.route('/<t_id>/<user_id>/cancel_set_sell', methods=["POST"])
@crossdomain(origin='*')
def cancel_set_sell(t_id, user_id):
	return check_stock_and_amount(t_id, user_id, request, Cmd.CANCEL_SET_SELL.value)


@app.route('/<t_id>/<user_id>/dumplog', methods=["GET"])
@crossdomain(origin='*')
def dumplog(t_id, user_id):
	filename = request.form["filename"]
	cmd = Command(TransactionID=int(t_id), C_type=Cmd.DUMPLOG.value, UserId=user_id,
				  Timestamp=time.time(), FileName=filename)

	if filename == "" and filename != "admin":
		return process_error(audit, cmd, "Parameter: 'filename' cannot be empty.")

	log = executor.submit(audit.UserCommand, *(cmd, ))
	resp = queue.PutItem(hash(user_id), cmd)
	log.result()
	return json.dumps(thrift_to_json(resp)) if type(resp) is Response else resp


def start_web_server():
	app.run(host=hostname, port=port)

if __name__ == "__main__":
	root = logging.getLogger()
	root.setLevel(logging.DEBUG)

	ch = logging.StreamHandler(sys.stdout)
	ch.setLevel(logging.DEBUG)
	formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s - [%(filename)s:%(lineno)s]')
	ch.setFormatter(formatter)
	root.addHandler(ch)

	start_web_server()
