package tools

import (
	"log"
	"time"

	"github.com/valyala/gorpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

const TxnServiceName = "TxnRPC"

type TxnConn interface {
	Send(cmd common.Command) *common.Response
	Close()
}

type txnServe struct {
	client   *gorpc.Client
	dispatch *gorpc.DispatcherClient
}

func (t *txnServe) Send(cmd common.Command) *common.Response {
	data, err := t.dispatch.Call(common.Commands[cmd.C_type], cmd)
	if err != nil {
		log.Println(err)
		return nil
	}
	resp, ok := data.(*common.Response)
	if !ok {
		log.Println("Failed to assert response type")
		return nil
	}
	return resp
}

func (t *txnServe) Close() {
	t.client.Stop()
}

func GetTxnConn() TxnConn {
	gorpc.RegisterType(&common.Command{})

	client := gorpc.NewTCPClient(common.CFG.TxnServer.Url)
	dispatcher := gorpc.NewDispatcher()
	dispatcher.AddService(TxnServiceName, &TxnRPC{})
	dispatchClient := dispatcher.NewServiceClient(TxnServiceName, client)
	client.Start()
	return &txnServe{client: client, dispatch: dispatchClient}
}

type TxnRPC struct {
	cache   CacheUtil
	session *CacheMongoSession
	logger  Logger
}

func (ts *TxnRPC) error(cmd *common.Command, msg string) (*common.Response, error) {
	log.Println("ERROR", msg)
	go ts.logger.ErrorEvent(cmd, msg)
	return &common.Response{Success: false, Message: msg}, nil
}

func (ts *TxnRPC) ADD(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetSharedInstance()
	defer db.Close()

	_, err := db.Users.AddUserMoney(cmd.UserId, cmd.Amount)
	if err != nil {
		return ts.error(cmd, "Failed to create and/or add money to account")
	}

	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "add", cmd.TransactionID)
	return &common.Response{Success: true}, nil
}

func (ts *TxnRPC) QUOTE(cmd *common.Command) (*common.Response, error) {
	data, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Quote server failed to respond with quote")
	}
	return &common.Response{Success: true, Quote: data.Quote, Stock: data.Symbol}, nil
}

func (ts *TxnRPC) BUY(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetSharedInstance()
	defer db.Close()

	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user "+cmd.UserId+" does not exist")
	}
	cacheReserve := ts.cache.GetReserved(cmd.UserId)
	if user.Balance-cacheReserve < cmd.Amount {
		return ts.error(cmd, "Specified amount is greater than can afford")
	}
	quote, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock")
	}

	shares := int(cmd.Amount / quote.Quote)
	if shares <= 0 {
		return ts.error(cmd, "Specified amount is not enough to purchase any shares")
	}
	cost := int64(shares) * quote.Quote
	expiry := time.Now().Add(time.Minute)

	pending := common.PendingTxn{UserId: cmd.UserId, Type: "BUY", Price: cost, Shares: shares,
		Reserved: cmd.Amount, Stock: quote.Symbol, Expiry: expiry}
	ts.cache.PushPendingTxn(pending)

	return &common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: cost, Shares: shares, Expiration: expiry.Unix()}, nil
}

func (ts *TxnRPC) COMMIT_BUY(cmd *common.Command) (*common.Response, error) {
	buy := ts.cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return ts.error(cmd, "There are no pending transactions")
	}

	db := ts.session.GetUniqueInstance()
	defer db.Close()

	_, err := db.Users.ProcessTxn(buy, true)
	if err != nil {
		return ts.error(cmd, "User can no longer afford this purchase")
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "remove", cmd.TransactionID)

	_, err = db.Transactions.LogTxn(buy, false)
	if err != nil {
		return ts.error(cmd, "Failed to store transaction log in database")
	}

	return &common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares, Paid: buy.Price}, nil
}

func (ts *TxnRPC) CANCEL_BUY(cmd *common.Command) (*common.Response, error) {
	buy := ts.cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return ts.error(cmd, "There is no buy to cancel")
	}
	return &common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares}, nil
}

func (ts *TxnRPC) SELL(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetSharedInstance()
	defer db.Close()

	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user "+cmd.UserId+" does not exist")
	} else if user.Stock[cmd.StockSymbol].Real == 0 {
		return ts.error(cmd, "User does not own any shares for that stock")
	}

	quote, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock")
	}
	actualShares := int(cmd.Amount / quote.Quote)
	shares := actualShares
	if shares <= 0 {
		return ts.error(cmd, "A single share is worth more than specified amount")
	} else if user.Stock[cmd.StockSymbol].Real < shares {
		shares = user.Stock[cmd.StockSymbol].Real
	}

	sellFor := int64(shares) * quote.Quote
	expiry := time.Now().Add(time.Minute)

	pending := common.PendingTxn{UserId: cmd.UserId, Type: "SELL", Price: sellFor, Shares: shares, Stock: quote.Symbol, Expiry: expiry}
	ts.cache.PushPendingTxn(pending)

	return &common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: int64(actualShares) * quote.Quote,
		Shares: actualShares, SharesAfford: shares, AffordAmount: sellFor, Expiration: expiry.Unix()}, nil
}

func (ts *TxnRPC) COMMIT_SELL(cmd *common.Command) (*common.Response, error) {
	sell := ts.cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return ts.error(cmd, "There are no pending transactions")
	}

	db := ts.session.GetUniqueInstance()
	defer db.Close()

	_, err := db.Users.ProcessTxn(sell, true)
	if err != nil {
		return ts.error(cmd, "User no longer has the correct number of shares to sell")
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "add", cmd.TransactionID)

	_, err = db.Transactions.LogTxn(sell, false)
	if err != nil {
		log.Println("!!IMPORTANT!! Failed to log sell")
	}

	return &common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares, Received: sell.Price}, nil
}

func (ts *TxnRPC) CANCEL_SELL(cmd *common.Command) (*common.Response, error) {
	sell := ts.cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return ts.error(cmd, "There is no sell to cancel")
	}
	return &common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares}, nil
}

func (ts *TxnRPC) SET_BUY_AMOUNT(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}
	cachedReserve := ts.cache.GetReserved(cmd.UserId)
	if user.Balance-cachedReserve < cmd.Amount {
		return ts.error(cmd, "Not enough funds")
	}
	_, err = ts.cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock")
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
	if _, err = db.Users.ReserveMoney(cmd.UserId, cmd.Amount); err != nil {
		return ts.error(cmd, "Failed to reserve even though should have")
	} else if _, err = db.Triggers.Set(trigger); err != nil {
		go db.Users.UnreserveMoney(cmd.UserId, cmd.Amount)
		return ts.error(cmd, "Failed to set trigger even though should have")
	}
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID)

	return &common.Response{Success: true}, nil
}

func (ts *TxnRPC) CANCEL_SET_BUY(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}

	trig, err := db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return ts.error(cmd, "No buy trigger to cancel")
	}
	_, err = db.Users.UnreserveMoney(cmd.UserId, trig.Amount)
	if err != nil {
		log.Println(err)
		return ts.error(cmd, "Internal server error")
	}
	go ts.logger.AccountTransaction(cmd.UserId, trig.Amount, "unreserve", cmd.TransactionID)

	return &common.Response{Success: true, Stock: cmd.StockSymbol}, nil
}

func (ts *TxnRPC) SET_BUY_TRIGGER(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}

	trig, err := db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return ts.error(cmd, "User must set buy amount first")
	}

	trig.When = cmd.Amount
	_, err = db.Triggers.Set(trig)
	if err != nil {
		return ts.error(cmd, "Internal error during operation")
	}

	return &common.Response{Success: true}, nil
}

func (ts *TxnRPC) SET_SELL_AMOUNT(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}
	realStocks := user.Stock[cmd.StockSymbol].Real - ts.cache.GetReservedShares(cmd.UserId)[cmd.StockSymbol]
	if realStocks <= 0 {
		return ts.error(cmd, "The user does not have any stock")
	}

	quote, err := ts.cache.GetQuote(cmd.StockSymbol, cmd.UserId, cmd.TransactionID)
	if err != nil {
		return ts.error(cmd, "Failed to get quote for that stock")
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

	_, err = db.Triggers.Set(trigger)
	if err != nil {
		return ts.error(cmd, "Failed to set sell amount")
	}
	db.Users.ReserveShares(cmd.UserId, cmd.StockSymbol, reservedShares)
	go ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "reserve", cmd.TransactionID)

	return &common.Response{Success: true}, nil
}

func (ts *TxnRPC) SET_SELL_TRIGGER(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}

	trig, err := db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return ts.error(cmd, "User must set sell amount first")
	}

	trig.When = cmd.Amount
	db.Triggers.Set(trig)

	return &common.Response{Success: true}, nil
}

func (ts *TxnRPC) CANCEL_SET_SELL(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	_, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}

	trig, err := db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return ts.error(cmd, "No sell trigger to cancel")
	}

	_, err = db.Users.UnreserveShares(cmd.UserId, cmd.StockSymbol, trig.Shares)
	if err != nil {
		log.Println(err)
		return ts.error(cmd, "Internal server error")
	}
	go ts.logger.AccountTransaction(cmd.UserId, trig.Amount, "unreserve", cmd.TransactionID)

	return &common.Response{Success: true}, nil
}

func (ts *TxnRPC) DUMPLOG(cmd *common.Command) (*common.Response, error) {
	var data *[]byte
	var err error
	if cmd.UserId != "admin" {
		db := ts.session.GetSharedInstance()
		defer db.Close()

		_, err = db.Users.GetUser(cmd.UserId)
		if err != nil {
			return ts.error(cmd, "The user does not exist")
		}
		data, err = ts.logger.DumpLogUser(cmd.UserId)
		if err != nil {
			log.Println(err)
			return ts.error(cmd, "Failed to get user log")
		}
	} else {
		data, err = ts.logger.DumpLog()
		if err != nil {
			return ts.error(cmd, "Failed to get log")
		}
	}
	return &common.Response{Success: true, File: data}, nil
}

func (ts *TxnRPC) DISPLAY_SUMMARY(cmd *common.Command) (*common.Response, error) {
	db := ts.session.GetUniqueInstance()
	defer db.Close()

	user, err := db.Users.GetUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "The user does not exist")
	}

	transactions, err := db.Transactions.Get(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "Failed to get transactions")
	}

	triggers, err := db.Triggers.GetAllUser(cmd.UserId)
	if err != nil {
		return ts.error(cmd, "Failed to get triggers")
	}

	cacheReserve := ts.cache.GetReserved(cmd.UserId)
	balance := user.Balance - cacheReserve
	reserved := user.Reserved + cacheReserve

	userCopy := *user
	cacheStocks := ts.cache.GetReservedShares(cmd.UserId)
	for k, v := range userCopy.Stock {
		v.Real = v.Real - cacheStocks[k]
		v.Reserved = v.Reserved + cacheStocks[k]
		userCopy.Stock[k] = v
	}

	return &common.Response{
		Success:      true,
		Status:       &common.UserInfo{Balance: balance, Reserved: reserved, Stock: userCopy.Stock},
		Transactions: &transactions.Logged,
		Triggers:     &triggers,
	}, nil
}

func GetTxnRPC(c CacheUtil, session *CacheMongoSession, l Logger) *TxnRPC {
	return &TxnRPC{cache: c, session: session, logger: l}
}
