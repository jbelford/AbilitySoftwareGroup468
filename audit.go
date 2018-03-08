package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	"github.com/mattpaletta/AbilitySoftwareGroup468/networks"
)

type AuditServer struct{}

func (ad *AuditServer) Start() {
	log.Println("Requesting RPC")
	logger, db := networks.GetLoggerRPC()
	defer db.Close()
	rpc.Register(logger)
	ln, err := net.Listen("tcp", common.CFG.AuditServer.Url)
	log.Println("connected to:", common.CFG.AuditServer.Url)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on:", common.CFG.AuditServer.Url)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		log.Println("Accepted connection")
		go rpc.ServeConn(conn)
	}
}
