package common

import (
	"bufio"
	"fmt"
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
		tcpConn, err := net.Dial("tcp", CFG.Quoteserver.Address)
		if err != nil {
			return nil, err
		}
		tcpConn.Write([]byte(fmt.Sprintf("%s, %s\n", symbol, userid)))
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
