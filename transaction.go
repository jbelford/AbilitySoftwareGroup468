package main

import (
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/networks"

	"github.com/mattpaletta/AbilitySoftwareGroup468/tools"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type TxnRPC struct {
	cache  tools.Cache
	db     *tools.MongoDB
	logger networks.Logger
}

func (ts *TxnRPC) error(cmd *common.Command, msg string, resp *common.Response) error {
	log.Println("ERROR", msg, msg)
	go ts.logger.ErrorEvent(cmd, msg)
	*resp = common.Response{Success: false, Message: msg}
	return nil
}

func (ts *TxnRPC) ADD(cmd *common.Command, resp *common.Response) error {
	err := ts.db.Users.AddUserMoney(cmd.UserId, cmd.Amount)
	if err != nil {
		return ts.error(cmd, "Failed to create and/or add money to account", resp)
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "add", cmd.TransactionID)
	*resp = common.Response{Success: true}
	return nil
}

func (ts *TxnRPC) QUOTE(cmd *common.Command, resp *common.Response) error {
	data, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Quote server failed to respond with quote", resp)
	}
	*resp = common.Response{Success: true, Quote: data.Quote, Stock: data.Symbol}
	return nil
}

func (ts *TxnRPC) BUY(cmd *common.Command, resp *common.Response) error {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user "+user.UserId+" does not exist", resp)
	}
	cacheReserve := ts.cache.GetReserved(cmd.UserId)
	if user.Balance-cacheReserve < cmd.Amount {
		return ts.error(cmd, "Specified amount is greater than can afford", resp)
	}
	quote, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock", resp)
	}

	shares := int(cmd.Amount / quote.Quote)
	if shares <= 0 {
		return ts.error(cmd, "Specified amount is not enough to purchase any shares", resp)
	}
	cost := int64(shares) * quote.Quote
	expiry := time.Now().Add(time.Minute)

	pending := common.PendingTxn{UserId: cmd.UserId, Type: "BUY", Price: cost, Shares: shares,
		Reserved: cmd.Amount, Stock: quote.Symbol, Expiry: expiry}
	ts.cache.PushPendingTxn(pending)

	*resp = common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: cost, Shares: shares, Expiration: expiry.Unix()}
	return nil
}

func (ts *TxnRPC) COMMIT_BUY(cmd *common.Command, resp *common.Response) error {
	buy := ts.cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return ts.error(cmd, "There are no pending transactions", resp)
	}

	err := ts.db.Users.ProcessTxn(buy, true)
	if err != nil {
		return ts.error(cmd, "User can no longer afford this purchase", resp)
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "remove", cmd.TransactionID)

	err = ts.db.Transactions.LogTxn(buy, false)
	if err != nil {
		return ts.error(cmd, "Failed to store transaction log in database", resp)
	}

	*resp = common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares, Paid: buy.Price}
	return nil
}

func (ts *TxnRPC) CANCEL_BUY(cmd *common.Command, resp *common.Response) error {
	buy := ts.cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return ts.error(cmd, "There is no buy to cancel", resp)
	}
	*resp = common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares}
	return nil
}

func (ts *TxnRPC) SELL(cmd *common.Command, resp *common.Response) error {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user "+user.UserId+" does not exist", resp)
	} else if user.Stock[cmd.StockSymbol].Real == 0 {
		return ts.error(cmd, "User does not own any shares for that stock", resp)
	}

	quote, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock", resp)
	}
	actualShares := int(cmd.Amount / quote.Quote)
	shares := actualShares
	if shares <= 0 {
		return ts.error(cmd, "A single share is worth more than specified amount", resp)
	} else if user.Stock[cmd.StockSymbol].Real < shares {
		shares = user.Stock[cmd.StockSymbol].Real
	}

	sellFor := int64(shares) * quote.Quote
	expiry := time.Now().Add(time.Minute)

	pending := common.PendingTxn{UserId: cmd.UserId, Type: "SELL", Price: sellFor, Shares: shares, Stock: quote.Symbol, Expiry: expiry}
	ts.cache.PushPendingTxn(pending)

	*resp = common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: int64(actualShares) * quote.Quote,
		Shares: actualShares, SharesAfford: shares, AffordAmount: sellFor, Expiration: expiry.Unix()}
	return nil
}

func (ts *TxnRPC) COMMIT_SELL(cmd *common.Command, resp *common.Response) error {
	sell := ts.cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return ts.error(cmd, "There are no pending transactions", resp)
	}

	err := ts.db.Users.ProcessTxn(sell, true)
	if err != nil {
		return ts.error(cmd, "User no longer has the correct number of shares to sell", resp)
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "add", cmd.TransactionID)

	err = ts.db.Transactions.LogTxn(sell, false)
	if err != nil {
		log.Println("!!IMPORTANT!! Failed to log sell")
	}

	*resp = common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares, Received: sell.Price}
	return nil
}

func (ts *TxnRPC) CANCEL_SELL(cmd *common.Command, resp *common.Response) error {
	sell := ts.cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return ts.error(cmd, "There is no sell to cancel", resp)
	}
	*resp = common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares}
	return nil
}

func (ts *TxnRPC) SET_BUY_AMOUNT(cmd *common.Command, resp *common.Response) error {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}
	cachedReserve := ts.cache.GetReserved(cmd.UserId)
	if user.Balance-cachedReserve < cmd.Amount {
		return ts.error(cmd, "Not enough funds", resp)
	}
	_, err = ts.cache.GetQuote(cmd.StockSymbol, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock", resp)
	}

	trigger := &common.Trigger{
		UserId:        cmd.UserId,
		Stock:         cmd.StockSymbol,
		TransactionID: cmd.TransactionID,
		Type:          "BUY",
		Amount:        cmd.Amount,
		When:          0,
	}
	// Reserve the money and then set the trigger
	if err = ts.db.Users.ReserveMoney(cmd.UserId, cmd.Amount); err != nil {
		return ts.error(cmd, "Failed to reserve even though should have", resp)
	} else if err = ts.db.Triggers.Set(trigger); err != nil {
		go ts.db.Users.UnreserveMoney(cmd.UserId, cmd.Amount)
		return ts.error(cmd, "Failed to set trigger even though should have", resp)
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID)

	*resp = common.Response{Success: true}
	return nil
}

func (ts *TxnRPC) CANCEL_SET_BUY(cmd *common.Command, resp *common.Response) error {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}

	trig, err := ts.db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return ts.error(cmd, "No buy trigger to cancel", resp)
	}
	err = ts.db.Users.UnreserveMoney(cmd.UserId, trig.Amount)
	if err != nil {
		log.Println(err)
		return ts.error(cmd, "Internal server error", resp)
	}
	go ts.logger.AccountTransaction(cmd.UserId, trig.Amount, "unreserve", cmd.TransactionID)

	*resp = common.Response{Success: true, Stock: cmd.StockSymbol}
	return nil
}

func (ts *TxnRPC) SET_BUY_TRIGGER(cmd *common.Command, resp *common.Response) error {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}

	trig, err := ts.db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return ts.error(cmd, "User must set buy amount first", resp)
	}

	trig.When = cmd.Amount
	err = ts.db.Triggers.Set(trig)
	if err != nil {
		return ts.error(cmd, "Internal error during operation", resp)
	}

	*resp = common.Response{Success: true}
	return nil
}

func (ts *TxnRPC) SET_SELL_AMOUNT(cmd *common.Command, resp *common.Response) error {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}
	realStocks := user.Stock[cmd.StockSymbol].Real - ts.cache.GetReservedShares(cmd.UserId)[cmd.StockSymbol]
	if realStocks <= 0 {
		return ts.error(cmd, "The user does not have any stock", resp)
	}

	quote, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock", resp)
	}

	// Get reserved shares
	reservedShares := int(cmd.Amount / quote.Quote)
	if reservedShares > realStocks {
		reservedShares = realStocks
	}

	trigger := &common.Trigger{
		UserId:        cmd.UserId,
		Type:          "SELL",
		TransactionID: cmd.TransactionID,
		Shares:        reservedShares,
		Stock:         cmd.StockSymbol,
		Amount:        cmd.Amount,
		When:          0,
	}

	err = ts.db.Triggers.Set(trigger)
	if err != nil {
		return ts.error(cmd, "Failed to set sell amount", resp)
	}
	ts.db.Users.ReserveShares(cmd.UserId, cmd.StockSymbol, reservedShares)
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID)

	*resp = common.Response{Success: true}
	return nil
}

func (ts *TxnRPC) SET_SELL_TRIGGER(cmd *common.Command, resp *common.Response) error {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}

	trig, err := ts.db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return ts.error(cmd, "User must set sell amount first", resp)
	}

	trig.When = cmd.Amount
	ts.db.Triggers.Set(trig)

	*resp = common.Response{Success: true}
	return nil
}

func (ts *TxnRPC) CANCEL_SET_SELL(cmd *common.Command, resp *common.Response) error {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}

	trig, err := ts.db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return ts.error(cmd, "No sell trigger to cancel", resp)
	}

	err = ts.db.Users.UnreserveShares(cmd.UserId, cmd.StockSymbol, trig.Shares)
	if err != nil {
		log.Println(err)
		return ts.error(cmd, "Internal server error", resp)
	}
	go ts.logger.AccountTransaction(cmd.UserId, trig.Amount, "unreserve", cmd.TransactionID)

	*resp = common.Response{Success: true}
	return nil
}

func (ts *TxnRPC) DUMPLOG(cmd *common.Command, resp *common.Response) error {
	var data *[]byte
	var err error
	if cmd.UserId != "admin" {
		_, err = ts.db.Users.GetUser(cmd.UserId)
		if err != nil {
			return ts.error(cmd, "The user does not exist", resp)
		}
		data, err = ts.logger.DumpLogUser(cmd.UserId)
		if err != nil {
			log.Println(err)
			return ts.error(cmd, "Failed to get user log", resp)
		}
	} else {
		data, err = ts.logger.DumpLog()
		if err != nil {
			return ts.error(cmd, "Failed to get log", resp)
		}
	}
	*resp = common.Response{Success: true, File: data}
	return nil
}

func (ts *TxnRPC) DISPLAY_SUMMARY(cmd *common.Command, resp *common.Response) error {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist", resp)
	}

	transactions, err := ts.db.Transactions.Get(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "Failed to get transactions", resp)
	}

	triggers, err := ts.db.Triggers.GetAllUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "Failed to get triggers", resp)
	}

	cacheReserve := ts.cache.GetReserved(cmd.UserId)
	balance := user.Balance - cacheReserve
	reserved := user.Reserved + cacheReserve

	cacheStocks := ts.cache.GetReservedShares(cmd.UserId)
	for k, v := range user.Stock {
		v.Real = v.Real - cacheStocks[k]
		v.Reserved = v.Reserved + cacheStocks[k]
		user.Stock[k] = v
	}

	*resp = common.Response{
		Success:      true,
		Status:       &common.UserInfo{Balance: balance, Reserved: reserved, Stock: user.Stock},
		Transactions: &transactions,
		Triggers:     &triggers,
	}
	return nil
}

type TransactionServer struct{}

func (ts *TransactionServer) Start() {
	logger := networks.GetLogger(common.CFG.TxnServer.Url)
	cache := tools.NewCache(logger)
	db := tools.GetMongoDatabase()
	defer db.Close()

	// Start trigger manager
	tm := tools.NewTrigMan(cache, db, logger)
	tm.Start()

	txn := &TxnRPC{cache: cache, db: db, logger: logger}
	rpc.Register(txn)

	ln, err := net.Listen("tcp", common.CFG.TxnServer.Url)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
