package main

import (
	"log"

	"github.com/valyala/gorpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/tools"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type TransactionServer struct{}

func (ts *TransactionServer) Start() {
	db := tools.NewCacheDB()
	defer db.Close()
	logger := tools.GetLogger(common.CFG.TxnServer.Url)
	defer logger.Close()
	cache := tools.NewCache(logger)

	// Start trigger manager
	tm := tools.NewTrigMan(cache, db, logger)
	tm.Start()

	txn := tools.GetTxnRPC(cache, db, logger)

	dispatcher := gorpc.NewDispatcher()
	dispatcher.AddService(tools.TxnServiceName, txn)
	server := gorpc.NewTCPServer(common.CFG.TxnServer.Url, dispatcher.NewHandlerFunc())

	err := server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
