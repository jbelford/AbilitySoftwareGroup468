package common

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
)

var qsConfig QuoteServConfig
var tcpConn net.Conn

func GetQuote(symbol string) (*QuoteData, error) {
	tcpConn.Write([]byte(symbol))
	msg, err := bufio.NewReader(tcpConn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	args := strings.Split(msg, ",")
	quote, _ := strconv.ParseFloat(args[0], 64)
	timestamp, _ := strconv.ParseUint(args[3], 10, 64)
	data := &QuoteData{
		Quote:     int(quote * 100),
		Symbol:    args[1],
		UserId:    args[2],
		Timestamp: timestamp,
		Cryptokey: args[4]}
	return data, nil
}

func init() {
	config, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	qsConfig = config.quoteserver
	conn, err := net.Dial("tcp", "quoteserve.seng:4443")
	if err != nil {
		log.Fatal(err)
	}
	tcpConn = conn
}
