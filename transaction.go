package main

import (
	"bufio"
	"encoding/json"
	"github.com/mattpaletta/AbilitySoftwareGroup468/logging"
	"log"
	"net"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type TransactionServer struct {
	cache  common.Cache
	db     *common.MongoDB
	logger logging.Logger
}

func (ts *TransactionServer) handle_add(cmd *common.Command) *common.Response {
	err := ts.db.Users.AddUserMoney(cmd.UserId, cmd.Amount)
	if err != nil {
		ts.logger.ErrorEvent(cmd, "Failed to create and/or add money to account")
		return &common.Response{Success: false, Message: "Failed to create and/or add money to account"}
	}
	err = ts.logger.AccountTransaction(cmd.UserId, cmd.Amount, "add")
	if err != nil {
		log.Println(err)
	}
	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_quote(cmd *common.Command) *common.Response {
	data, err := ts.cache.GetQuote(cmd.StockSymbol)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed"}
	}
	return &common.Response{Success: true, Quote: data.Quote, Stock: data.Symbol}
}

func (ts *TransactionServer) handle_buy(cmd *common.Command) *common.Response {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "User does not exist"}
	}
	cacheReserve := ts.cache.GetReserved(user.UserId)
	if user.Balance-cacheReserve < cmd.Amount {
		return &common.Response{Success: false, Message: "Specified amount is greater than can afford"}
	}
	quote, err := ts.cache.GetQuote(cmd.StockSymbol)
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
	ts.cache.PushPendingTxn(pending)

	return &common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: cost, Shares: shares, Expiration: expiry.Unix()}
}

func (ts *TransactionServer) handle_commit_buy(cmd *common.Command) *common.Response {
	buy := ts.cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return &common.Response{Success: false, Message: "There are no pending transactions"}
	}

	err := ts.db.Users.ProcessBuy(buy)
	if err != nil {
		return &common.Response{Success: false, Message: "User can no longer afford this purchase"}
	}
	err = ts.db.Transactions.LogTxn(buy, false)
	if err != nil {
		log.Println("Failed to log buy")
	}

	return &common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares, Paid: buy.Price}
}

func (ts *TransactionServer) handle_cancel_buy(cmd *common.Command) *common.Response {
	buy := ts.cache.PopPendingTxn(cmd.UserId, "BUY")
	if buy == nil {
		return &common.Response{Success: false, Message: "There is no buy to cancel"}
	}
	return &common.Response{Success: true, Stock: buy.Stock, Shares: buy.Shares}
}

func (ts *TransactionServer) handle_sell(cmd *common.Command) *common.Response {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "User does not exist"}
	} else if user.Stock[cmd.StockSymbol].Real == 0 {
		return &common.Response{Success: false, Message: "User does not own any shares for that stock"}
	}

	quote, err := ts.cache.GetQuote(cmd.StockSymbol)
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
	ts.cache.PushPendingTxn(pending)

	return &common.Response{Success: true, ReqAmount: cmd.Amount, RealAmount: int64(actualShares) * quote.Quote,
		Shares: actualShares, SharesAfford: shares, AffordAmount: sellFor, Expiration: expiry.Unix()}
}

func (ts *TransactionServer) handle_commit_sell(cmd *common.Command) *common.Response {
	sell := ts.cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return &common.Response{Success: false, Message: "There are no pending transactions"}
	}

	err := ts.db.Users.ProcessSell(sell)
	if err != nil {
		return &common.Response{Success: false, Message: "User no longer has the correct number of shares to sell"}
	}
	err = ts.db.Transactions.LogTxn(sell, false)
	if err != nil {
		log.Println("Failed to log sell")
	}

	return &common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares, Received: sell.Price}
}

func (ts *TransactionServer) handle_cancel_sell(cmd *common.Command) *common.Response {
	sell := ts.cache.PopPendingTxn(cmd.UserId, "SELL")
	if sell == nil {
		return &common.Response{Success: false, Message: "There is no sell to cancel"}
	}
	return &common.Response{Success: true, Stock: sell.Stock, Shares: sell.Shares}
}

func (ts *TransactionServer) handle_set_buy_amount(cmd *common.Command) *common.Response {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	} else if user.Balance < cmd.Amount {
		return &common.Response{Success: false, Message: "Not enough funds"}
	}
	_, err = ts.cache.GetQuote(cmd.StockSymbol)
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
	ts.db.Triggers.Set(trigger)
	ts.db.Users.ReserveMoney(cmd.UserId, cmd.Amount)

	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_cancel_set_buy(cmd *common.Command) *common.Response {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := ts.db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return &common.Response{Success: false, Message: "No buy trigger to cancel"}
	}
	err = ts.db.Users.UnreserveMoney(cmd.UserId, trig.Amount)
	if err != nil {
		log.Println(err)
		return &common.Response{Success: false, Message: "Internal server error"}
	}

	return &common.Response{Success: true, Stock: cmd.StockSymbol}
}

func (ts *TransactionServer) handle_set_buy_trigger(cmd *common.Command) *common.Response {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := ts.db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "BUY")
	if err != nil {
		return &common.Response{Success: false, Message: "User must set buy amount first"}
	}

	trig.When = cmd.Amount
	err = ts.db.Triggers.Set(trig)
	if err != nil {
		return &common.Response{Success: false, Message: "Internal error during operation"}
	}

	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_set_sell_amount(cmd *common.Command) *common.Response {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}
	realStocks := user.Stock[cmd.StockSymbol].Real - ts.cache.GetReservedShares(cmd.UserId, cmd.StockSymbol)
	if realStocks == 0 {
		return &common.Response{Success: false, Message: "The user does not have any stock"}
	}

	quote, err := ts.cache.GetQuote(cmd.StockSymbol)
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

	err = ts.db.Triggers.Set(trigger)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to set sell amount"}
	}
	ts.db.Users.ReserveShares(cmd.UserId, cmd.StockSymbol, reservedShares)

	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_set_sell_trigger(cmd *common.Command) *common.Response {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := ts.db.Triggers.Get(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return &common.Response{Success: false, Message: "User must set sell amount first"}
	}

	trig.When = cmd.Amount
	ts.db.Triggers.Set(trig)

	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_cancel_set_sell(cmd *common.Command) *common.Response {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	trig, err := ts.db.Triggers.Cancel(cmd.UserId, cmd.StockSymbol, "SELL")
	if err != nil {
		return &common.Response{Success: false, Message: "No sell trigger to cancel"}
	}

	err = ts.db.Users.UnreserveShares(cmd.UserId, cmd.StockSymbol, trig.Shares)
	if err != nil {
		log.Println(err)
		return &common.Response{Success: false, Message: "Internal server error"}
	}

	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_admin_dumplog(cmd *common.Command) *common.Response {
	log.Println("handle_admin_dumplog")
	//success
	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_dumplog(cmd *common.Command) *common.Response {
	_, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	return &common.Response{Success: true}
}

func (ts *TransactionServer) handle_display_summary(cmd *common.Command) *common.Response {
	user, err := ts.db.Users.GetUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "The user does not exist"}
	}

	transactions, err := ts.db.Transactions.Get(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to get transactions"}
	}

	triggers, err := ts.db.Triggers.GetAllUser(cmd.UserId)
	if err != nil {
		return &common.Response{Success: false, Message: "Failed to get triggers"}
	}

	return &common.Response{
		Success:      true,
		Status:       &common.UserInfo{Balance: user.Balance, Reserved: user.Reserved, Stock: user.Stock},
		Transactions: &transactions,
		Triggers:     &triggers,
	}
}

func (ts *TransactionServer) Start() {
	ts.logger = logging.GetLogger(common.CFG.TxnServer.Url)
	ts.cache = common.NewCache()
	mongoDb, err := common.GetMongoDatabase()
	if err != nil {
		log.Fatal(err)
	}
	ts.db = mongoDb

	tm := common.NewTrigMan(ts.cache, ts.db)
	tm.Start()

	defer ts.db.Close()
	ln, err := net.Listen("tcp", common.CFG.TxnServer.Url)
	if err != nil {
		log.Fatal(err)
	}

	handler := common.NewCommandHandler()

	handler.On(common.ADD, ts.handle_add)
	handler.On(common.QUOTE, ts.handle_quote)
	handler.On(common.BUY, ts.handle_buy)
	handler.On(common.COMMIT_BUY, ts.handle_commit_buy)
	handler.On(common.CANCEL_BUY, ts.handle_cancel_buy)
	handler.On(common.SELL, ts.handle_sell)
	handler.On(common.COMMIT_SELL, ts.handle_commit_sell)
	handler.On(common.CANCEL_SELL, ts.handle_cancel_sell)
	handler.On(common.SET_BUY_AMOUNT, ts.handle_set_buy_amount)
	handler.On(common.CANCEL_SET_BUY, ts.handle_cancel_set_buy)
	handler.On(common.SET_BUY_TRIGGER, ts.handle_set_buy_trigger)
	handler.On(common.SET_SELL_AMOUNT, ts.handle_set_sell_amount)
	handler.On(common.SET_SELL_TRIGGER, ts.handle_set_sell_trigger)
	handler.On(common.CANCEL_SET_SELL, ts.handle_cancel_set_sell)
	handler.On(common.DUMPLOG, ts.handle_dumplog)
	handler.On(common.ADMIN_DUMPLOG, ts.handle_admin_dumplog)
	handler.On(common.DISPLAY_SUMMARY, ts.handle_display_summary)

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
