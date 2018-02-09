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
	logger, writer := networks.GetLoggerRPC()
	defer writer.Close()
	rpc.Register(logger)
	ln, err := net.Listen("tcp", common.CFG.AuditServer.Url)
	logger, writer := networks.GetLoggerRPC()
	log.Println("connected to:", common.CFG.AuditServer.Url)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on:", common.CFG.AuditServer.Url)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go rpc.ServeConn(conn)
	}
}
