package main

import (
	"bufio"
	"log"
	"net"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

var db *common.MongoDB

type TransactionServer struct{}

func handle_add(cmd common.Command) {
	err := db.AddUserMoney(cmd.UserId, cmd.Amount)
	log.Fatal(err)
}

func handle_quote(cmd common.Command) {
	log.Println("handle_quote")
}

func handle_buy(cmd common.Command) {
	log.Println("handle_buy")
}

func handle_commit_buy(cmd common.Command) {
	log.Println("handle_commit_buy")
}

func handle_cancel_buy(cmd common.Command) {
	log.Println("handle_cancel_buy")
}

func handle_sell(cmd common.Command) {
	log.Println("handle_sell")
}

func handle_commit_sell(cmd common.Command) {
	log.Println("handle_commit_sell")
}

func handle_cancel_sell(cmd common.Command) {
	log.Println("handle_cancel_sell")
}

func handle_set_buy_amount(cmd common.Command) {
	log.Println("handle_set_buy_amount")
}

func handle_cancel_set_buy(cmd common.Command) {
	log.Println("handle_cancel_set_buy")
}

func handle_set_buy_trigger(cmd common.Command) {
	log.Println("handle_set_buy_trigger")
}

func handle_set_sell_amount(cmd common.Command) {
	log.Println("handle_set_sell_amount")
}

func handle_set_sell_trigger(cmd common.Command) {
	log.Println("handle_set_sell_trigger")
}

func handle_cancel_set_sell(cmd common.Command) {
	log.Println("handle_cancel_set_sell")
}

func handle_admin_dumplog(cmd common.Command) {
	log.Println("handle_admin_dumplog")
}

func handle_dumplog(cmd common.Command) {
	log.Println("handle_dumplog")
}

func handle_display_summary(cmd common.Command) {
	log.Println("handle_display_summary")
}

func (ts *TransactionServer) Start() {
	mongoDb, err := common.GetMongoDatabase()
	if err != nil {
		log.Fatal(err)
	}
	db = mongoDb
	defer db.Close()
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")
	handler := common.CommandHandler{}

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
		message, _ := bufio.NewReader(conn).ReadString('\n')
		log.Println("Received: ", string(message))
		handler.Parse(message)

	}
}
