package main

import (
	"github.com/mattpaletta/AbilitySoftwareGroup468/logging"
	"log"
	"net"
	"net/rpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type AuditServer struct{}

func (ad *AuditServer) Start() {
	logger, writer := logging.GetLoggerRPC()
	defer writer.Close()
	rpc.Register(logger)
	ln, err := net.Listen("tcp", common.CFG.AuditServer.Url)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		rpc.ServeConn(conn)
	}
}
