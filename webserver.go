package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	"github.com/mattpaletta/AbilitySoftwareGroup468/tools"
)

type WebServer struct {
	logger  tools.Logger
	txnConn tools.TxnConn
}

func (ws *WebServer) error(cmd *common.Command, msg string) *common.Response {
	log.Println(msg)
	go ws.logger.ErrorEvent(cmd, msg)
	return &common.Response{Success: false, Message: msg}
}

func (ws *WebServer) Start() {
	ws.txnConn = tools.GetTxnConn()
	defer ws.txnConn.Close()
	ws.logger = tools.GetLogger(common.CFG.WebServer.LUrl)
	defer ws.logger.Close()

	var dir string
	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", ws.indexHandler).Methods("GET")

	r.HandleFunc("/{t_id}/{user_id}/display_summary", wrapHandler(ws.userSummaryHandler)).Methods("GET")

	r.HandleFunc("/{t_id}/{user_id}/add", wrapHandler(ws.userAddHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/quote", wrapHandler(ws.userQuoteHandler)).Methods("GET")

	//buying stocks
	r.HandleFunc("/{t_id}/{user_id}/buy", wrapHandler(ws.userBuyHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/commit_buy", wrapHandler(ws.userCommitBuyHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/cancel_buy", wrapHandler(ws.userCancelBuyHandler)).Methods("POST")

	//selling stocks
	r.HandleFunc("/{t_id}/{user_id}/sell", wrapHandler(ws.userSellHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/commit_sell", wrapHandler(ws.userCommitSellHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/cancel_sell", wrapHandler(ws.userCancelSellHandler)).Methods("POST")

	//buy triggers
	r.HandleFunc("/{t_id}/{user_id}/set_buy_amount", wrapHandler(ws.userSetBuyAmountHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/cancel_set_buy", wrapHandler(ws.userCancelSetBuyHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/set_buy_trigger", wrapHandler(ws.userSetBuyTriggerHandler)).Methods("POST")

	//sell triggers
	r.HandleFunc("/{t_id}/{user_id}/set_sell_amount", wrapHandler(ws.userSetSellAmountHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/set_sell_trigger", wrapHandler(ws.userSetSellTriggerHandler)).Methods("POST")
	r.HandleFunc("/{t_id}/{user_id}/cancel_set_sell", wrapHandler(ws.userCancelSetSellHandler)).Methods("POST")

	//user log
	r.HandleFunc("/{t_id}/{user_id}/dumplog", wrapHandler(ws.userDumplogHandler)).Methods("GET")

	r.Handler(http.FileServer(http.Dir(dir))))

	log.Println("Listening on:", common.CFG.WebServer.LUrl)
	srv := &http.Server{
		Handler: r,
		Addr:    common.CFG.WebServer.LUrl,
	}

	log.Fatal(srv.ListenAndServe())
}

/*
Handles basic page visibility function
returns page template
*/
func (ws *WebServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("test.html")
	t, _ = t.ParseFiles("./test.html")
	t.Execute(w, "")
}

/*
	displays user information

	```json
	{
	  "success": true,
	  "status": {
	    "balance": 2000
	  },
	  "transactions": [
	    {
	      "type": "BUY",
	      "triggered": false,
	      "stock": "ABC",
	      "amount": 192.15,
	      "shares": 20,
	      "timestamp": 1516767552619
	    }
	  ],
	  "triggers": [
	    {
	      "stock": "ABC",
	      "type": "SELL",
	      "amount": 200,
	      "when": 10.50
	    }
	  ]
	}
	```
*/
func (ws *WebServer) userSummaryHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.DISPLAY_SUMMARY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserSummary")
	}
	return resp
}

/*
	Handles adding cash to user accounts

	JSON return
	```json
	{
	"success": true
	}
	```
*/
func (ws *WebServer) userAddHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.ADD,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Count not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserAdd")
	}
	return resp
}

/*
		Handler for users requesting quote price

		JSON return
		```json
		{
		"success": true,
		"stock": "ABC",
		"quote": 12.50
	}
	```
*/
func (ws *WebServer) userQuoteHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.QUOTE,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserQuote")
	}
	return resp
}

/*
		buys an amount of stock

		JSON Response
		```json
		{
		"success": true,
		"amount_requested": 200,
		"real_amount": 195.15,
		"shares": 20,
		"expiration": 1516767552619
	}
	```
*/
func (ws *WebServer) userBuyHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.BUY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Could not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserBuy")
	}
	return resp
}

/*
	Default handler, for any url that does not require validity testing
	commit buy, cancel buy, commit sell, cancel sell,
*/
func (ws *WebServer) userCommitBuyHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.COMMIT_BUY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserCommitBuy")
	}
	return resp
}

/*

	cancel buy
*/
func (ws *WebServer) userCancelBuyHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.CANCEL_BUY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserCancelBuy")
	}
	return resp
}

/*
sells an amount of stocks

JSON response
```json
{
  "success": true,
  "amount_requested": 200,
  "real_amount": 182.21,
  "shares": 19,
  "expiration": 1516767552619
}
```
*/
func (ws *WebServer) userSellHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.SELL,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Could not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserSell")
	}
	return resp
}

/*
	commit sell
*/
func (ws *WebServer) userCommitSellHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.COMMIT_SELL,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserCommitSell")
	}
	return resp
}

/*
	cancel sell
*/
func (ws *WebServer) userCancelSellHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.CANCEL_SELL,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserCancelSell")
	}
	return resp
}

/*
sets a buy amount for a trigger

json response
```json
{
  "success": true
}
```
*/
func (ws *WebServer) userSetBuyAmountHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.SET_BUY_AMOUNT,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Could not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserSetBuyAmount")
	}
	return resp
}

/*
cancels previous set buys

```json
{
  "success": true,
  "stock": "ABC"
}
```
*/
func (ws *WebServer) userCancelSetBuyHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.CANCEL_SET_BUY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserCancelSetBuy")
	}
	return resp
}

/*
sets buy triggers

```json
{
  "success": true
}
```
*/
func (ws *WebServer) userSetBuyTriggerHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.SET_BUY_TRIGGER,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Could not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserSetBuyTrigger")
	}
	return resp
}

/*
sell
JSON response

```json
{
  "success": true
}
```
*/
func (ws *WebServer) userSetSellAmountHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.SET_SELL_AMOUNT,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Could not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserSetSellAmount")
	}
	return resp
}

/*
sell
JSON response

```json
{
  "success": true
}
```
*/
func (ws *WebServer) userSetSellTriggerHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.SET_SELL_TRIGGER,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	var err error
	cmd.Amount, err = strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return ws.error(&cmd, "Could not process field: 'amount'")
	} else if cmd.Amount <= 0 {
		return ws.error(&cmd, "Parameter: 'amount' must be greater than 0")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserSetSellTrigger")
	}
	return resp
}

/*
sell
JSON response

```json
{
  "success": true
}
```
*/
func (ws *WebServer) userCancelSetSellHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		TransactionID: t_id,
		C_type:        common.CANCEL_SET_SELL,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	cmd.StockSymbol = r.URL.Query().Get("stock")
	if cmd.StockSymbol == "" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'stock' cannot be an empty string")
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserCancelSetSell")
	}
	return resp
}

/*
dumps a log
*/
func (ws *WebServer) userDumplogHandler(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response {
	cmd := common.Command{
		FileName:      r.URL.Query().Get("filename"),
		TransactionID: t_id,
		UserId:        mux.Vars(r)["user_id"],
		C_type:        common.DUMPLOG,
		Timestamp:     time.Now(),
	}
	if cmd.FileName == "" && cmd.UserId != "admin" { //should maybe do is alpha numeric check here
		return ws.error(&cmd, "Parameter: 'filename' cannot be an empty string")
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ws.error(&cmd, "Internal error prevented operation: UserDumpLog")
	} else if resp.Success {
		w.Header().Set("Content-Disposition", "attachment; filename="+cmd.FileName)
		w.Header().Set("Content-Type", "application/xml")
		io.Copy(w, bytes.NewReader(*resp.File))
	}
	return resp
}

func wrapHandler(
	handler func(w http.ResponseWriter, r *http.Request, t_id int64) *common.Response,
) func(w http.ResponseWriter, r *http.Request) {

	h := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request: " + r.URL.EscapedPath())

		w.Header().Set("Content-Type", "application/json")
		// test input here/validity of requester
		t_id, err := strconv.ParseInt(mux.Vars(r)["t_id"], 10, 64)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp := handler(w, r, t_id)

		if w.Header().Get("Content-Type") == "application/json" {
			respJSON, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Write(respJSON)
			}
		}
	}
	return h
}
