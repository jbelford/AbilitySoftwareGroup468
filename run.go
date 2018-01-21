package main

import (
	"os"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	"github.com/mattpaletta/AbilitySoftwareGroup468/transaction"
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
		server = new(transaction.TransactionServer)
	}
	// addr = os.Args[1]
	// port = os.Args[2]
}

func main() {
	server.Start()
	// // Start server
	// address := fmt.Sprintf("%s:%s", addr, port)
	// log.Printf("Listening at %s\n", address)
	// listener, err := net.Listen("tcp", address)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Accept connection to the port
	// conn, err := listener.Accept()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// msg, _ := bufio.NewReader(conn).ReadString('\n')
	// log.Printf("Received message: %s", msg)
}
