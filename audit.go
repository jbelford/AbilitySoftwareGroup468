package main

import (
	"log"

	"github.com/valyala/gorpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	"github.com/mattpaletta/AbilitySoftwareGroup468/tools"
)

type AuditServer struct{}

func (ad *AuditServer) Start() {
	log.Println("Requesting RPC")
	session := tools.GetMongoSession()
	defer session.Close()

	dispatcher := gorpc.NewDispatcher()

	logger := tools.GetLoggerRPC(session)
	dispatcher.AddService(tools.LoggerServiceName, logger)

	server := gorpc.NewTCPServer(common.CFG.AuditServer.Url, dispatcher.NewHandlerFunc())
	log.Println("connected to:", common.CFG.AuditServer.Url)

	err := server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
