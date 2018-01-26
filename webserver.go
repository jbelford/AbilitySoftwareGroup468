package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type WebServer struct{}

func (ws *WebServer) Start() {
	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")

	r.HandleFunc("/{user_id}/display_summary", wrapHandler(userSummaryHandler)).Methods("GET")

	r.HandleFunc("/{user_id}/add", wrapHandler(userAddHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/quote", wrapHandler(userQuoteHandler)).Methods("GET")

	//buying stocks
	r.HandleFunc("/{user_id}/buy", wrapHandler(userBuyHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/commit_buy", wrapHandler(userCommitBuyHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_buy", wrapHandler(userCancelBuyHandler)).Methods("POST")

	//selling stocks
	r.HandleFunc("/{user_id}/sell", wrapHandler(userSellHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/commit_sell", wrapHandler(userCommitSellHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_sell", wrapHandler(userCancelSellHandler)).Methods("POST")

	//buy triggers
	r.HandleFunc("/{user_id}/set_buy_amount", wrapHandler(userSetBuyAmountHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_set_buy", wrapHandler(userCancelSetBuyHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/set_buy_trigger", wrapHandler(userSetBuyTriggerHandler)).Methods("POST")

	//sell triggers
	r.HandleFunc("/{user_id}/set_sell_amount", wrapHandler(userSetSellAmountHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/set_sell_trigger", wrapHandler(userSetSellTriggerHandler)).Methods("POST")
	r.HandleFunc("/{user_id}/cancel_set_sell", wrapHandler(userCancelSetSellHandler)).Methods("POST")

	//user log
	r.HandleFunc("/{user_id}/dumplog", wrapHandler(userDumplogHandler)).Methods("POST")

	//admin log
	r.HandleFunc("/{admin_id}/dumplog", wrapHandler(adminDumplogHandler)).Methods("POST")

	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(dir))))

	srv := &http.Server{
		Handler:      r,
		Addr:         common.CFG.WebServer.Url,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

/*
Handles basic page visibility function
returns page template
*/
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// user_info := common.CommandConstructor(r.FormValue("data"))
	// passInfo(user_info)
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
func userSummaryHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		C_type:    common.DISPLAY_SUMMARY,
		UserId:    mux.Vars(r)["user_id"],
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userAddHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return &common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return &common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}
	cmd := common.Command{
		C_type:    common.ADD,
		UserId:    mux.Vars(r)["user_id"],
		Amount:    amount,
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userQuoteHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}
	cmd := common.Command{
		C_type:      common.QUOTE,
		UserId:      mux.Vars(r)["user_id"],
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
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
		C_type:      common.BUY,
		UserId:      mux.Vars(r)["user_id"],
		Amount:      amount,
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

/*
	Default handler, for any url that does not require validity testing
	commit buy, cancel buy, commit sell, cancel sell,
*/
func userCommitBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		C_type:    common.COMMIT_BUY,
		UserId:    mux.Vars(r)["user_id"],
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

/*

	cancel buy
*/
func userCancelBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		C_type:    common.CANCEL_BUY,
		UserId:    mux.Vars(r)["user_id"],
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
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
		C_type:      common.SELL,
		UserId:      mux.Vars(r)["user_id"],
		Amount:      amount,
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

/*
	commit sell
*/
func userCommitSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		C_type:    common.COMMIT_SELL,
		UserId:    mux.Vars(r)["user_id"],
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

/*
	cancel sell
*/
func userCancelSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	cmd := common.Command{
		C_type:    common.CANCEL_SELL,
		UserId:    mux.Vars(r)["user_id"],
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userSetBuyAmountHandler(w http.ResponseWriter, r *http.Request) *common.Response {
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
		C_type:      common.SET_BUY_AMOUNT,
		UserId:      mux.Vars(r)["user_id"],
		Amount:      amount,
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userCancelSetBuyHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	cmd := common.Command{
		C_type:      common.CANCEL_SET_BUY,
		UserId:      mux.Vars(r)["user_id"],
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userSetBuyTriggerHandler(w http.ResponseWriter, r *http.Request) *common.Response {
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
		C_type:      common.SET_BUY_TRIGGER,
		UserId:      mux.Vars(r)["user_id"],
		Amount:      amount,
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userSetSellAmountHandler(w http.ResponseWriter, r *http.Request) *common.Response {
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
		C_type:      common.SET_SELL_AMOUNT,
		UserId:      mux.Vars(r)["user_id"],
		Amount:      amount,
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userSetSellTriggerHandler(w http.ResponseWriter, r *http.Request) *common.Response {
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
		C_type:      common.SET_SELL_TRIGGER,
		UserId:      mux.Vars(r)["user_id"],
		Amount:      amount,
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
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
func userCancelSetSellHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	quote_id := r.URL.Query().Get("stock")
	if quote_id == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	cmd := common.Command{
		C_type:      common.CANCEL_SET_SELL,
		UserId:      mux.Vars(r)["user_id"],
		StockSymbol: quote_id,
		Timestamp:   time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

/*
dumps a log
*/
func userDumplogHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	filename := r.URL.Query().Get("filename")
	if filename == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'filename' cannot be an empty string"}
	}

	cmd := common.Command{
		C_type:    common.DUMPLOG,
		FileName:  filename,
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

/*
dumps the big log
*/
func adminDumplogHandler(w http.ResponseWriter, r *http.Request) *common.Response {
	filename := r.URL.Query().Get("filename")
	if filename == "" { //should maybe do is alpha numeric check here
		return &common.Response{Success: false, Message: "Parameter: 'stock' cannot be an empty string"}
	}

	cmd := common.Command{
		C_type:    common.ADMIN_DUMPLOG,
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return &common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

//passes command to transaction server
func issueTransactionCommand(com common.Command) *common.Response {
	textCmd, err := json.Marshal(com)
	if err != nil {
		return nil
	}

	conn, err := net.Dial("tcp", "transaction.prod.ability.com:44421")
	if err != nil {
		return nil
	}
	defer conn.Close()

	_, err = conn.Write(append(textCmd, '\n'))

	var resp string
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	resp, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil
	}

	var jsonResp *common.Response
	err = json.Unmarshal([]byte(resp), &jsonResp)
	if err != nil {
		return nil
	}
	return jsonResp
}

func wrapHandler(
	handler func(w http.ResponseWriter, r *http.Request) *common.Response,
) func(w http.ResponseWriter, r *http.Request) {

	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// test input here/validity of requester
		resp := handler(w, r)

		respJSON, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write(respJSON)
		}
	}
	return h
}
