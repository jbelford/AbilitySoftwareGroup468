package main

import (
	"os"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

var server common.Server
var addr string
var port string

func init() {
	if len(os.Args) < 2 {
		panic("Missing arguments: <cmdLine>")
	}
	servType := os.Args[1]
	switch servType {
	case "transaction":
		server = new(TransactionServer)
	case "webserver":
		server = new(WebServer)
	}
}

func main() {
	server.Start()
}
