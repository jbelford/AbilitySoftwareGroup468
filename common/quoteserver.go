package common

import (
	"fmt"
	"log"
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
		body := []byte(fmt.Sprintf("%s, %s\n", symbol, userid))
		data, err := ClusterDialTCP(CFG.Quoteserver.Address, body, 3)
		if err != nil {
			return nil, err
		}
		msg = string(data[:])
	}
	log.Println(msg)
	args := strings.Split(msg, ",")
	for i, a := range args {
		args[i] = strings.TrimSpace(a)
	}
	quote, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		log.Println("GetQuote: Failed to parse price data")
		return nil, err
	}
	timestamp, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		log.Println("GetQuote: Failed to parse timestamp")
		return nil, err
	}
	data := &QuoteData{
		Quote:     int64(quote * 100),
		Symbol:    args[1],
		UserId:    args[2],
		Timestamp: timestamp,
		Cryptokey: args[4]}
	return data, nil
}
