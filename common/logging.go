package common

import (
	"encoding/xml"
  "time"
)

type Args struct {
  Action, Command, Cryptokey, DebugMessage, ErrorMessage, FileName, Server, StockSymbol, Username string
  QuoteServerTime, TransactionNum int
  Funds, Price int64
  Timestamp uint64
}

type UserCommand struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		Timestamp       time.Time `xml:"timestamp"`
  		TransactionNum  string    `xml:"transactionNum"`
  		Command         string    `xml:"command"`
  		Username        string    `xml:"username, omitempty"`
  		StockSymbol     string    `xml:"stockSymbol, omitempty"`
  		Filename        string    `xml:"filename, omitempty"`
  		Funds           string    `xml:"funds, omitempty"`
}

type QuoteServer struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		Timestamp       time.Time `xml:"timestamp"`
  		TransactionNum  string    `xml:"transactionNum"`
  		Price           string    `xml:"price"`
  		StockSymbol     string    `xml:"stockSymbol"`
  		Username        string    `xml:"username"`
  		QuoteServerTime string    `xml:"quoteServerTime"`
  		Cryptokey       string    `xml:"cryptokey"`
}

type AccountTransaction struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		Timestamp       time.Time `xml:"timestamp"`
  		TransactionNum  string    `xml:"transactionNum"`
      Action          string    `xml:"action"`
      Username        string    `xml:"username"`
      Funds           string    `xml:"funds"`
}

type SystemEvent struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		Timestamp       time.Time `xml:"timestamp"`
  		TransactionNum  string    `xml:"transactionNum"`
  		Command         string    `xml:"command"`
  		Username        string    `xml:"username, omitempty"`
  		StockSymbol     string    `xml:"stockSymbol, omitempty"`
  		Filename        string    `xml:"filename, omitempty"`
  		Funds           string    `xml:"funds, omitempty"`
}

type ErrorEvent struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		Timestamp       time.Time `xml:"timestamp"`
  		TransactionNum  string    `xml:"transactionNum"`
  		Command         string    `xml:"command"`
  		Username        string    `xml:"username, omitempty"`
  		StockSymbol     string    `xml:"stockSymbol, omitempty"`
  		Filename        string    `xml:"filename, omitempty"`
  		Funds           string    `xml:"funds, omitempty"`
  		ErrorMessage    string    `xml:"errorMessage, omitempty"`
}

type DebugEvent struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		Timestamp       time.Time `xml:"timestamp"`
  		TransactionNum  string    `xml:"transactionNum"`
  		Command         string    `xml:"command"`
  		Username        string    `xml:"username, omitempty"`
  		StockSymbol     string    `xml:"stockSymbol, omitempty"`
  		Filename        string    `xml:"filename, omitempty"`
  		Funds           string    `xml:"funds, omitempty"`
  		DebugMessage    string    `xml:"debugMessage, omitempty"`
}
