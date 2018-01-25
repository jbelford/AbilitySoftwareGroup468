package common

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

var qsConfig QuoteServConfig
var tcpConn net.Conn

func GetQuote(symbol string) (*QuoteData, error) {
	var msg string
	if qsConfig.Mock {
		msg = fmt.Sprintf("12.50,%s,NA,1111111111,123198fadfa", symbol)
	} else {
		tcpConn.Write([]byte(symbol))
		var err error
		msg, err = bufio.NewReader(tcpConn).ReadString('\n')
		if err != nil {
			return nil, err
		}
	}
	args := strings.Split(msg, ",")
	quote, _ := strconv.ParseFloat(args[0], 64)
	timestamp, _ := strconv.ParseUint(args[3], 10, 64)
	data := &QuoteData{
		Quote:     int64(quote * 100),
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
	qsConfig = config.Quoteserver
	if !qsConfig.Mock {
		tcpConn, err = net.Dial("tcp", qsConfig.Address)
		if err != nil {
			log.Fatal(err)
		}
	}
}
