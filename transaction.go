package main

import (
	"log"

	"github.com/valyala/gorpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/tools"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type TransactionServer struct{}

func (ts *TransactionServer) Start() {
	logger := tools.GetLogger(common.CFG.TxnServer.LUrl)
	defer logger.Close()

	util := tools.NewCacheUtil(logger)
	session := tools.GetCacheMongoSession()
	defer session.Close()

	// Start trigger manager
	tm := tools.NewTrigMan(util, session, logger)
	tm.Start()

	txn := tools.GetTxnRPC(util, session, logger)

	dispatcher := gorpc.NewDispatcher()
	dispatcher.AddService(tools.TxnServiceName, txn)
	server := gorpc.NewTCPServer(common.CFG.TxnServer.LUrl, dispatcher.NewHandlerFunc())

	err := server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
