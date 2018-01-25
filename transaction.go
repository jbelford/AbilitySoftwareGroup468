package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

var db *common.MongoDB

type TransactionServer struct{}

func handle_add(cmd *common.Command) *common.Response {
	err := db.AddUserMoney(cmd.UserId, cmd.Amount)
	if err != nil {
		log.Println(err)
		return &common.Response{Success: false, Message: "Failed"}
	}
	return &common.Response{Success: true}
}

func handle_quote(cmd *common.Command) *common.Response {
	log.Println("handle_quote")
	// TODO:// Get the current quote from quote cache, return in response.
	// RESPONSE: success, stock, quote
	return nil
}

func handle_buy(cmd *common.Command) *common.Response {
	log.Println("handle_buy")
	// TODO://
	// RESPONSE: success, amount_requested, real_amount, shares, expiration
	return nil
}

func handle_commit_buy(cmd *common.Command) *common.Response {
	log.Println("handle_commit_buy")
	//success, stock, shares, paid
	return nil
}

func handle_cancel_buy(cmd *common.Command) *common.Response {
	log.Println("handle_cancel_buy")
	// success, stock, shares
	return nil
}

func handle_sell(cmd *common.Command) *common.Response {
	log.Println("handle_sell")
	// success, amount_requested, real_amount, shares, expiration
	return nil
}

func handle_commit_sell(cmd *common.Command) *common.Response {
	log.Println("handle_commit_sell")
	// success, stock, shares, received
	return nil
}

func handle_cancel_sell(cmd *common.Command) *common.Response {
	log.Println("handle_cancel_sell")
	//success, stock, shares
	return nil
}

func handle_set_buy_amount(cmd *common.Command) *common.Response {
	log.Println("handle_set_buy_amount")
	//success
	return nil
}

func handle_cancel_set_buy(cmd *common.Command) *common.Response {
	log.Println("handle_cancel_set_buy")
	// success, stock
	return nil
}

func handle_set_buy_trigger(cmd *common.Command) *common.Response {
	log.Println("handle_set_buy_trigger")
	// success
	return nil
}

func handle_set_sell_amount(cmd *common.Command) *common.Response {
	log.Println("handle_set_sell_amount")
	// success
	return nil
}

func handle_set_sell_trigger(cmd *common.Command) *common.Response {
	log.Println("handle_set_sell_trigger")
	// success
	return nil
}

func handle_cancel_set_sell(cmd *common.Command) *common.Response {
	log.Println("handle_cancel_set_sell")
	// success
	return nil
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

func (ts *TransactionServer) Start() {
	mongoDb, err := common.GetMongoDatabase()
	if err != nil {
		log.Fatal(err)
	}
	db = mongoDb
	defer db.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:8081")
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
