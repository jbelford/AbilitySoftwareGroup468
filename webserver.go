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
	"sync"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/networks"

	"github.com/gorilla/mux"
	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type WebServer struct {
	logger  networks.Logger
	txnConn networks.TxnConn
}

func (ws *WebServer) error(cmd *common.Command, msg string) *common.Response {
	go ws.logger.ErrorEvent(cmd, msg)
	return &common.Response{Success: false, Message: msg}
}

type Counter struct {
	mu sync.Mutex
	x  int64
}

func (c *Counter) Inc() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.x = c.x + 1
	return c.x
}

var t_id Counter

func (ws *WebServer) Start() {
	ws.logger = networks.GetLogger(common.CFG.WebServer.Url)
	ws.txnConn = networks.GetTxnConn()
	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", ws.indexHandler).Methods("GET")

	r.HandleFunc("/{user_id}/display_summary", wrapHandler(ws.userSummaryHandler)).Methods("GET")

	r.HandleFunc("/{user_id}/add", wrapHandler(ws.userAddHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/quote", wrapHandler(ws.userQuoteHandler)).Methods("GET")

	//buying stocks
	r.HandleFunc("/{user_id}/buy", wrapHandler(ws.userBuyHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/commit_buy", wrapHandler(ws.userCommitBuyHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_buy", wrapHandler(ws.userCancelBuyHandler)).Methods("POST")

	//selling stocks
	r.HandleFunc("/{user_id}/sell", wrapHandler(ws.userSellHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/commit_sell", wrapHandler(ws.userCommitSellHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_sell", wrapHandler(ws.userCancelSellHandler)).Methods("POST")

	//buy triggers
	r.HandleFunc("/{user_id}/set_buy_amount", wrapHandler(ws.userSetBuyAmountHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_set_buy", wrapHandler(ws.userCancelSetBuyHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/set_buy_trigger", wrapHandler(ws.userSetBuyTriggerHandler)).Methods("POST")

	//sell triggers
	r.HandleFunc("/{user_id}/set_sell_amount", wrapHandler(ws.userSetSellAmountHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/set_sell_trigger", wrapHandler(ws.userSetSellTriggerHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_set_sell", wrapHandler(ws.userCancelSetSellHandler)).Methods("POST")

	//user log
	r.HandleFunc("/{user_id}/dumplog", wrapHandler(ws.userDumplogHandler)).Methods("GET")

	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(dir))))

	log.Println("Listening on:", common.CFG.WebServer.Url)
	srv := &http.Server{
		Handler: r,
		Addr:    common.CFG.WebServer.Url,
	}

	log.Fatal(srv.ListenAndServe())
}

/*
Handles basic page visibility function
returns page template
*/
func (ws *WebServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("test.html")
	t, _ = t.ParseFiles("./templates/test.html")
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
func (ws *WebServer) userSummaryHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.DISPLAY_SUMMARY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userAddHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.ADD,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userQuoteHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.QUOTE,
		UserId:        mux.Vars(r)["user_id"],
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.BUY,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	}
	return resp
}

/*
	Default handler, for any url that does not require validity testing
	commit buy, cancel buy, commit sell, cancel sell,
*/
func (ws *WebServer) userCommitBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.COMMIT_BUY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	}
	return resp
}

/*

	cancel buy
*/
func (ws *WebServer) userCancelBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.CANCEL_BUY,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.SELL,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	}
	return resp
}

/*
	commit sell
*/
func (ws *WebServer) userCommitSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.COMMIT_SELL,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	}
	return resp
}

/*
	cancel sell
*/
func (ws *WebServer) userCancelSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.CANCEL_SELL,
		UserId:        mux.Vars(r)["user_id"],
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userSetBuyAmountHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.SET_BUY_AMOUNT,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userCancelSetBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.CANCEL_SET_BUY,
		UserId:        mux.Vars(r)["user_id"],
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userSetBuyTriggerHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.SET_BUY_TRIGGER,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userSetSellAmountHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.SET_SELL_AMOUNT,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userSetSellTriggerHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	// amounts will be passed to the web server as a long to prevent any floating point conversion issues of any kind
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.SET_SELL_TRIGGER,
		UserId:        mux.Vars(r)["user_id"],
		Amount:        amount,
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
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
func (ws *WebServer) userCancelSetSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	cmd := common.Command{
		TransactionID: t_id.Inc(),
		C_type:        common.CANCEL_SET_SELL,
		UserId:        mux.Vars(r)["user_id"],
		StockSymbol:   quote_id,
		Timestamp:     time.Now(),
	}
	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	}
	return resp
}

/*
dumps a log
*/
func (ws *WebServer) userDumplogHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	userId := mux.Vars(r)["user_id"]
	filename := r.URL.Query().Get("filename")
	if filename == "" && userId != "admin" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'filename' cannot be an empty string"}
	}

	cmd := common.Command{
		FileName:      filename,
		TransactionID: t_id.Inc(),
		UserId:        mux.Vars(r)["user_id"],
		C_type:        common.DUMPLOG,
		Timestamp:     time.Now(),
	}

	go ws.logger.UserCommand(&cmd)

	resp := ws.txnConn.Send(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if resp.Success {
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", "application/xml")
		io.Copy(w, bytes.NewReader(*resp.File))
	}
	return resp
}

func wrapHandler(
	handler func(w http.ResponseWriter, r *http.Request) *common.Response,
) func(w http.ResponseWriter, r *http.Request) {

	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// test input here/validity of requester
		resp := handler(w, r)

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
