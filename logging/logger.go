package logging

import (
	"encoding/xml"
	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
	"io/ioutil"
	"net/rpc"
	"os"
	"time"
)

const (
	userCommandMethod        = "LoggerRPC.UserCommand"
	quoteServerMethod        = "LoggerRPC.QuoteServer"
	accountTransactionMethod = "LoggerRPC.AccountTransaction"
	systemEventMethod        = "LoggerRPC.SystemEvent"
	errorEventMethod         = "LoggerRPC.ErrorEvent"
	debugEventMethod         = "LoggerRPC.DebugEvent"
	dumpLogMethod            = "LoggerRPC.DumpLog"
)

type Logger interface {
	UserCommand(cmd *common.Command) error
	QuoteServer(quote *common.QuoteData, tid int64) error
	AccountTransaction(userId string, funds int64, action string, tid int64) error
	SystemEvent(cmd *common.Command) error
	ErrorEvent(cmd *common.Command, e string) error
	DebugEvent(cmd *common.Command, debug string) error
	DumpLogUser(userId string) (*[]byte, error)
	DumpLog() (*[]byte, error)
}

type logger struct {
	client *rpc.Client
	server string
}

func (l *logger) UserCommand(cmd *common.Command) error {
	args := &Args{
		TransactionNum: cmd.TransactionID,
		Timestamp:      uint64(time.Now().Unix()),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
	}
	err := l.client.Call(userCommandMethod, args, nil)
	return err
}

func (l *logger) QuoteServer(quote *common.QuoteData, tid int64) error {
	args := &Args{
		TransactionNum:  tid,
		Timestamp:       uint64(time.Now().Unix()),
		Server:          l.server,
		Username:        quote.UserId,
		Price:           quote.Quote,
		StockSymbol:     quote.Symbol,
		QuoteServerTime: quote.Timestamp,
		Cryptokey:       quote.Cryptokey,
	}
	err := l.client.Call(quoteServerMethod, args, nil)
	return err
}

func (l *logger) AccountTransaction(userId string, funds int64, action string, tid int64) error {
	args := &Args{
		TransactionNum: tid,
		Timestamp:      uint64(time.Now().Unix()),
		Server:         l.server,
		Action:         action,
		Username:       userId,
		Funds:          funds,
	}
	err := l.client.Call(accountTransactionMethod, args, nil)
	return err
}

func (l *logger) SystemEvent(cmd *common.Command) error {
	args := &Args{
		Timestamp:      uint64(time.Now().Unix()),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
		TransactionNum: cmd.TransactionID,
	}
	err := l.client.Call(systemEventMethod, args, nil)
	return err
}

func (l *logger) ErrorEvent(cmd *common.Command, e string) error {
	args := &Args{
		Timestamp:      uint64(time.Now().Unix()),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
		TransactionNum: cmd.TransactionID,
		ErrorMessage:   e,
	}
	err := l.client.Call(errorEventMethod, args, nil)
	return err
}

func (l *logger) DebugEvent(cmd *common.Command, debug string) error {
	args := &Args{
		Timestamp:      uint64(time.Now().Unix()),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
		DebugMessage:   debug,
		TransactionNum: cmd.TransactionID,
	}
	err := l.client.Call(debugEventMethod, args, nil)
	return err
}

func (l *logger) DumpLogUser(userId string) (*[]byte, error) {
	var data *[]byte
	err := l.client.Call(dumpLogMethod, &Args{FileName: userId + ".xml"}, &data)
	return data, err
}

func (l *logger) DumpLog() (*[]byte, error) {
	var data *[]byte
	err := l.client.Call(dumpLogMethod, &Args{}, &data)
	return data, err
}

func GetLogger(server string) Logger {
	client, err := rpc.Dial("tcp", common.CFG.AuditServer.Url)
	if err != nil {
		return nil
	}
	return &logger{client, server}
}

type LoggerRPC struct {
	writer *os.File
}

func (l *LoggerRPC) writeLogs(log interface{}, userFilename string) error {
	flag := os.O_APPEND
	if _, err := os.Stat(userFilename); os.IsNotExist(err) {
		flag |= os.O_CREATE
	}
	uWriter, err := os.OpenFile(userFilename, flag, 0600)
	if err != nil {
		return err
	}
	defer uWriter.Close()
	toWrite, err := xml.MarshalIndent(log, "  ", "    ")
	if err == nil {
		l.writer.Write(toWrite)
		uWriter.Write(toWrite)
	}
	return err
}

func (l *LoggerRPC) readLog(filename string) *[]byte {
	data := []byte("<?xml version=\"1.0\"?>\n<log>\n")
	read, err := ioutil.ReadFile(filename)
	if err != nil {
		data = append(data, read...)
	}
	data = append(data, []byte("\n</log>")...)
	return &data
}

func (l *LoggerRPC) UserCommand(args *Args, result *string) error {
	log := &UserCommand{
		Command:        args.Command,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	return l.writeLogs(log, args.Username+".xml")
}

func (l *LoggerRPC) QuoteServer(args *Args, result *string) error {
	log := &QuoteServer{
		Cryptokey:       args.Cryptokey,
		Price:           args.Price,
		QuoteServerTime: args.QuoteServerTime,
		StockSymbol:     args.StockSymbol,
		Timestamp:       args.Timestamp,
		TransactionNum:  args.TransactionNum,
		Username:        args.Username,
	}
	return l.writeLogs(log, args.Username+".xml")
}

func (l *LoggerRPC) AccountTransaction(args *Args, result *string) error {
	log := &AccountTransaction{
		Action:         args.Action,
		Funds:          args.Funds,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	return l.writeLogs(log, args.Username+".xml")
}

func (l *LoggerRPC) SystemEvent(args *Args, result *string) error {
	log := &SystemEvent{
		Command:        args.Command,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	return l.writeLogs(log, args.Username+".xml")
}

func (l *LoggerRPC) ErrorEvent(args *Args, result *string) error {
	log := &ErrorEvent{
		Command:        args.Command,
		ErrorMessage:   args.ErrorMessage,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	return l.writeLogs(log, args.Username+".xml")
}

func (l *LoggerRPC) DebugEvent(args *Args, result *string) error {
	log := &DebugEvent{
		Command:        args.Command,
		DebugMessage:   args.DebugMessage,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	return l.writeLogs(log, args.Username+".xml")
}

func (l *LoggerRPC) DumpLog(args *Args, result *[]byte) error {
	filename := "tmp.xml"
	if len(args.FileName) == 0 {
		filename = args.FileName
	}
	result = l.readLog(filename)
	return nil
}

func GetLoggerRPC() (*LoggerRPC, *os.File) {
	flag := os.O_APPEND
	if _, err := os.Stat("tmp.xml"); os.IsNotExist(err) {
		flag |= os.O_CREATE
	}
	writer, err := os.OpenFile("tmp.xml", flag, 0600)
	if err != nil {
		panic(err)
	}
	return &LoggerRPC{writer}, writer
}
