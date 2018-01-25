package common

import (
	"encoding/xml"
)

type Args struct {
  action, command, cryptokey, debugMessage, errorMessage, fileName, server, stockSymbol, timestamp, username string
  funds, price, quoteServerTime, transactionNum int
}

type UserCommand struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		timestamp       string    `xml:"timestamp"`
  		transactionNum  string    `xml:"transactionNum"`
  		command         string    `xml:"command"`
  		username        string    `xml:"username, omitempty"`
  		stockSymbol     string    `xml:"stockSymbol, omitempty"`
  		filename        string    `xml:"filename, omitempty"`
  		funds           string    `xml:"funds, omitempty"`
}

type QuoteServer struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		timestamp       string    `xml:"timestamp"`
  		transactionNum  string    `xml:"transactionNum"`
  		price           string    `xml:"price"`
  		stockSymbol     string    `xml:"stockSymbol"`
  		username        string    `xml:"username"`
  		quoteServerTime string    `xml:"quoteServerTime"`
  		cryptokey       string    `xml:"cryptokey"`
}

type AccountTransaction struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		timestamp       string    `xml:"timestamp"`
  		transactionNum  string    `xml:"transactionNum"`
      action          string    `xml:"action"`
      username        string    `xml:"username"`
      funds           string    `xml:"funds"`
}

type SystemEvent struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		timestamp       string    `xml:"timestamp"`
  		transactionNum  string    `xml:"transactionNum"`
  		command         string    `xml:"command"`
  		username        string    `xml:"username, omitempty"`
  		stockSymbol     string    `xml:"stockSymbol, omitempty"`
  		filename        string    `xml:"filename, omitempty"`
  		funds           string    `xml:"funds, omitempty"`
}

type ErrorEvent struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		timestamp       string    `xml:"timestamp"`
  		transactionNum  string    `xml:"transactionNum"`
  		command         string    `xml:"command"`
  		username        string    `xml:"username, omitempty"`
  		stockSymbol     string    `xml:"stockSymbol, omitempty"`
  		filename        string    `xml:"filename, omitempty"`
  		price           string    `xml:"price"`
  		errorMessage    string    `xml:"errorMessage, omitempty"`
}

type DebugEvent struct {
  		XMLName         xml.Name  `xml:"serCommand"`
  		timestamp       string    `xml:"timestamp"`
  		transactionNum  string    `xml:"transactionNum"`
  		command         string    `xml:"command"`
  		username        string    `xml:"username, omitempty"`
  		stockSymbol     string    `xml:"stockSymbol, omitempty"`
  		filename        string    `xml:"filename, omitempty"`
  		debugMessage    string    `xml:"debugMessage, omitempty"`
}
