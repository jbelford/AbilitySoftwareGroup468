package common

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// GetQuote returns quote data provided by the legacy quote server
func GetQuote(symbol string, userid string) (*QuoteData, error) {
	var msg string

	if CFG.Quoteserver.Mock {
		time.Sleep(time.Millisecond * 300)
		msg = fmt.Sprintf("12.50,%s,%s,1111111111,123198fadfa\n", symbol, userid)
	} else {
		log.Printf("QuoteServer: Requesting quote '%s'\n", symbol)
		tcpConn, err := net.DialTimeout("tcp", CFG.Quoteserver.Address, time.Second*5)
		if err != nil {
			return nil, err
		}
		defer tcpConn.Close()
		_, err = tcpConn.Write([]byte(fmt.Sprintf("%s, %s\n", symbol, userid)))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		msg, err = bufio.NewReader(tcpConn).ReadString('\n')
		if err != nil {
			return nil, err
		}
	}

	args := strings.Split(msg, ",")
	for i, a := range args {
		args[i] = strings.TrimSpace(a)
	}
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
