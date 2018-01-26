package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

var cache common.Cache
var db *common.MongoDB
var serverName string

type TransactionServer struct{}

func handle_add(cmd *common.Command) *common.Response {
	err := db.Users.AddUserMoney(cmd.UserId, cmd.Amount)
	if err != nil {
		log.Println(err)
		return &common.Response{Success: false, Message: "Failed"}
	}
	return &common.Response{Success: true}
}

func handle_quote(cmd *common.Command) *common.Response {
	data, err := cache.GetQuote(cmd.StockSymbol)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed"}
	}
	return &common.Response{Success: true, Quote: data.Quote, Stock: data.Symbol}
}

func handle_buy(cmd *common.Command) *common.Response {
	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "User does not exist"}
	}
	cacheReserve := cache.GetReserved(user.UserId)
	if user.Balance-cacheReserve < cmd.Amount {
		return &common.Response{Success: false, Message: "Specified amount is greater than can afford"}
	}
	quote, err := cache.GetQuote(cmd.StockSymbol)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to get quote for that stock"}
	}

	shares := int(cmd.Amount / quote.Quote)
	if shares <= 0 {
		return &common.Response{Success: false, Message: "Specified amount is not enough to purchase any shares"}
	}
	cost := int64(shares) * quote.Quote
	expiry := time.Now().Add(time.Minute)

	pending := common.PendingTxn{UserId: cmd.UserId, Type: "BUY", Price: cost, Shares: shares,
		Reserved: cmd.Amount, Stock: quote.Symbol, Expiry: expiry}
	cache.PushPendingTxn(pending)

	return &common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: cost, Shares: shares, Expiration: expiry.Unix()}
}

func handle_commit_buy(cmd *common.Command) *common.Response {
	buy := cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return &common.Response{Success: false, Message: "There are no pending transactions"}
	}

	err := db.Users.ProcessBuy(buy)
	if err != nil {
		return &common.Response{Success: false, Message: "User can no longer afford this purchase"}
	}

	return &common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares, Paid: buy.Price}
}

func handle_cancel_buy(cmd *common.Command) *common.Response {
	buy := cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return &common.Response{Success: false, Message: "There is no buy to cancel"}
	}
	return &common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares}
}

func handle_sell(cmd *common.Command) *common.Response {
	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "User does not exist"}
	} else if user.Stock[cmd.StockSymbol].Real == 0 {
		return &common.Response{Success: false, Message: "User does not own any shares for that stock"}
	}

	quote, err := cache.GetQuote(cmd.StockSymbol)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to get quote for that stock"}
	}
	actualShares := int(cmd.Amount / quote.Quote)
	shares := actualShares
	if shares <= 0 {
		return &common.Response{Success: false, Message: "A single share is worth more than specified amount"}
	} else if user.Stock[cmd.StockSymbol].Real < shares {
		shares = user.Stock[cmd.StockSymbol].Real
	}

	sellFor := int64(shares) * quote.Quote
	expiry := time.Now().Add(time.Minute)

	pending := common.PendingTxn{UserId: cmd.UserId, Type: "SELL", Price: sellFor, Shares: shares, Stock: quote.Symbol, Expiry: expiry}
	cache.PushPendingTxn(pending)

	return &common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: int64(actualShares) * quote.Quote,
		Shares: actualShares, SharesAfford: shares, AffordAmount: sellFor, Expiration: expiry.Unix()}
}

func handle_commit_sell(cmd *common.Command) *common.Response {
	sell := cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return &common.Response{Success: false, Message: "There are no pending transactions"}
	}

	err := db.Users.ProcessSell(sell)
	if err != nil {
		return &common.Response{Success: false, Message: "User no longer has the correct number of shares to sell"}
	}

	return &common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares, Received: sell.Price}
}

func handle_cancel_sell(cmd *common.Command) *common.Response {
	sell := cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return &common.Response{Success: false, Message: "There is no sell to cancel"}
	}
	return &common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares}
}

func handle_set_buy_amount(cmd *common.Command) *common.Response {
	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	} else if user.Balance < cmd.Amount {
		return &common.Response{Success: false, Message: "Not enough funds"}
	}
	_, err = cache.GetQuote(cmd.StockSymbol)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to get quote for that stock"}
	}

	trigger := &common.Trigger{
		UserId: cmd.UserId,
		Stock:  cmd.StockSymbol,
		Type:   "BUY",
		Amount: cmd.Amount,
		When:   0,
	}
	db.Triggers.Set(trigger)
	db.Users.ReserveMoney(cmd.UserId, cmd.Amount)

	return &common.Response{Success: true}
}

func handle_cancel_set_buy(cmd *common.Command) *common.Response {
	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return &common.Response{Success: false, Message: "No buy trigger to cancel"}
	}
	err = db.Users.UnreserveMoney(cmd.UserId, trig.Amount)
	if err != nil {
		log.Println(err)
		return &common.Response{Success: false, Message: "Internal server error"}
	}

	return &common.Response{Success: true, Stock: cmd.StockSymbol}
}

func handle_set_buy_trigger(cmd *common.Command) *common.Response {
	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return &common.Response{Success: false, Message: "User must set buy amount first"}
	}

	trig.When = cmd.Amount
	err = db.Triggers.Set(trig)
	if err != nil {
		return &common.Response{Success: false, Message: "Internal error during operation"}
	}

	return &common.Response{Success: true}
}

func handle_set_sell_amount(cmd *common.Command) *common.Response {
	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}
	realStocks := user.Stock[cmd.StockSymbol].Real - cache.GetReservedShares(cmd.UserId, cmd.StockSymbol)
	if realStocks == 0 {
		return &common.Response{Success: false, Message: "The user does not have any stock"}
	}

	quote, err := cache.GetQuote(cmd.StockSymbol)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to get quote for that stock"}
	}

	// Get reserved shares
	reservedShares := int(cmd.Amount / quote.Quote)
	if reservedShares > realStocks {
		reservedShares = realStocks
	}

	trigger := &common.Trigger{
		UserId: cmd.UserId,
		Type:   "SELL",
		Shares: reservedShares,
		Stock:  cmd.StockSymbol,
		Amount: cmd.Amount,
		When:   0,
	}

	err = db.Triggers.Set(trigger)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to set sell amount"}
	}
	db.Users.ReserveShares(cmd.UserId, cmd.StockSymbol, reservedShares)

	return &common.Response{Success: true}
}

func handle_set_sell_trigger(cmd *common.Command) *common.Response {
	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return &common.Response{Success: false, Message: "User must set sell amount first"}
	}

	trig.When = cmd.Amount
	db.Triggers.Set(trig)

	return &common.Response{Success: true}
}

func handle_cancel_set_sell(cmd *common.Command) *common.Response {
	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return &common.Response{Success: false, Message: "No sell trigger to cancel"}
	}

	err = db.Users.UnreserveShares(cmd.UserId, cmd.StockSymbol, trig.Shares)
	if err != nil {
		log.Println(err)
		return &common.Response{Success: false, Message: "Internal server error"}
	}

	return &common.Response{Success: true}
}

func handle_admin_dumplog(cmd *common.Command) *common.Response {
	log.Println("handle_admin_dumplog")
	//success
	return nil
}

func handle_dumplog(cmd *common.Command) *common.Response {
	log.Println("handle_dumplog")
	// success
	return nil
}

func handle_display_summary(cmd *common.Command) *common.Response {
	log.Println("handle_display_summary")
	// success, status{balance}, transactions[{type, triggered, stock, amount, shares, timestamp}], triggers[{stock, type, amount, when}]
	return nil
}

func createUserCommandLog(cmd *common.Command, tranNum int) *common.Args {

	args := &common.Args{
		Timestamp:      uint64(cmd.Timestamp.Unix()),
		Server:         serverName,
		TransactionNum: tranNum,
		Username:       cmd.UserId,
		Funds:          cmd.Amount,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
	}
	return args
}

func createQuoteServerLog(quote *common.QuoteData, tranNum int) *common.Args {

	args := &common.Args{
		Timestamp:      quote.Timestamp,
		Server:         serverName,
		TransactionNum: tranNum,
		Username:       quote.UserId,
		Price:          quote.Quote,
		StockSymbol:    quote.Symbol,
		Cryptokey:      quote.Cryptokey,
	}
	return args
}

func createAccountTransactionLog(cmd *common.Command, tranNum int, action string) *common.Args {

	timestamp := time.Now()
	args := &common.Args{
		Timestamp:      uint64(timestamp.Unix()),
		Server:         serverName,
		TransactionNum: tranNum,
		Action:         action,
		Username:       cmd.UserId,
		Funds:          cmd.Amount,
	}
	return args
}

func LogResult(args common.Args, logtype string) {
	client, err := rpc.Dial("tcp", "auditserver.prod.ability.com:44422")
	if err != nil {
		log.Fatal(err)
	}
	var result Result
	err = client.Call(logtype, args, &result)

	//Do we care about getting anything back from the audit server?
}

func (ts *TransactionServer) Start() {
	cache = common.NewCache()
	mongoDb, err := common.GetMongoDatabase()
	if err != nil {
		log.Fatal(err)
	}
	db = mongoDb

	tm := common.NewTrigMan(cache, db)
	tm.Start()

	defer db.Close()
	ln, err := net.Listen("tcp", "transaction.prod.ability.com:44421")
	if err != nil {
		log.Fatal(err)
	}

	handler := common.NewCommandHandler()

	handler.On(common.ADD, handle_add)
	handler.On(common.QUOTE, handle_quote)
	handler.On(common.BUY, handle_buy)
	handler.On(common.COMMIT_BUY, handle_commit_buy)
	handler.On(common.CANCEL_BUY, handle_cancel_buy)
	handler.On(common.SELL, handle_sell)
	handler.On(common.COMMIT_SELL, handle_commit_sell)
	handler.On(common.CANCEL_SELL, handle_cancel_sell)
	handler.On(common.SET_BUY_AMOUNT, handle_set_buy_amount)
	handler.On(common.CANCEL_SET_BUY, handle_cancel_set_buy)
	handler.On(common.SET_BUY_TRIGGER, handle_set_buy_trigger)
	handler.On(common.SET_SELL_AMOUNT, handle_set_sell_amount)
	handler.On(common.SET_SELL_TRIGGER, handle_set_sell_trigger)
	handler.On(common.CANCEL_SET_SELL, handle_cancel_set_sell)
	handler.On(common.DUMPLOG, handle_dumplog)
	handler.On(common.ADMIN_DUMPLOG, handle_admin_dumplog)
	handler.On(common.DISPLAY_SUMMARY, handle_display_summary)

	for {
		conn, err := ln.Accept()
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			continue
		}
		log.Println("Received: ", string(message))
		var resp *common.Response
		resp, err = handler.Parse(message)
		if err != nil {
			log.Println(err)
			resp = &common.Response{Success: false, Message: "Internal error parsing request"}
		}
		var respByte []byte
		respByte, err = json.Marshal(resp)
		conn.Write(append(respByte, '\n'))
		conn.Close()
	}
}
