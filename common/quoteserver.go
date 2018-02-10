package common

import (
	"bufio"
	"net"
	"strconv"
	"strings"
)

func GetQuote(symbol string) (*QuoteData, error) {
	var msg string

	tcpConn, err := net.Dial("tcp", CFG.Quoteserver.Address)
	if err != nil {
		return nil, err
	}
	tcpConn.Write([]byte(symbol + "\n"))
	msg, err = bufio.NewReader(tcpConn).ReadString('\n')
	if err != nil {
		return nil, err
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
