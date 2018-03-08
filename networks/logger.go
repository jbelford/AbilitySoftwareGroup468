package networks

import (
	"encoding/xml"
	"log"
	"net/rpc"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

const (
	userCommandMethod = "LoggerRPC.UserCommand"
	quoteServerMethod = "LoggerRPC.QuoteServer"
	accountTxnMethod  = "LoggerRPC.AccountTransaction"
	systemEventMethod = "LoggerRPC.SystemEvent"
	errorEventMethod  = "LoggerRPC.ErrorEvent"
	debugEventMethod  = "LoggerRPC.DebugEvent"
	dumpLogMethod     = "LoggerRPC.DumpLog"
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
	Close() error
}

type logger struct {
	client *rpc.Client
	server string
}

func (l *logger) Call(method string, args interface{}, result interface{}) (err error) {
	for {
		err = l.client.Call(method, args, result)
		if err == rpc.ErrShutdown {
			l.client, err = rpc.Dial("tcp", common.CFG.AuditServer.Url)
			continue
		}
		return err
	}
}

func (l *logger) UserCommand(cmd *common.Command) error {
	args := &UserCommand{
		TransactionNum: cmd.TransactionID,
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		Funds:          cmd.Amount,
	}
	return l.Call(userCommandMethod, args, nil)
}

func (l *logger) QuoteServer(quote *common.QuoteData, tid int64) error {
	args := &QuoteServer{
		TransactionNum:  tid,
		Timestamp:       uint64(time.Now().Unix() * 1000),
		Server:          l.server,
		Username:        quote.UserId,
		Price:           quote.Quote,
		StockSymbol:     quote.Symbol,
		QuoteServerTime: quote.Timestamp,
		Cryptokey:       quote.Cryptokey,
	}
	return l.Call(quoteServerMethod, args, nil)
}

func (l *logger) AccountTransaction(userId string, funds int64, action string, tid int64) error {
	args := &AccountTransaction{
		TransactionNum: tid,
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Action:         action,
		Username:       userId,
		Funds:          funds,
	}
	return l.Call(accountTxnMethod, args, nil)
}

func (l *logger) SystemEvent(cmd *common.Command) error {
	args := &SystemEvent{
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		Funds:          cmd.Amount,
		TransactionNum: cmd.TransactionID,
	}
	return l.Call(systemEventMethod, args, nil)
}

func (l *logger) ErrorEvent(cmd *common.Command, e string) error {
	args := &ErrorEvent{
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		Funds:          cmd.Amount,
		TransactionNum: cmd.TransactionID,
		ErrorMessage:   e,
	}
	return l.Call(errorEventMethod, args, nil)
}

func (l *logger) DebugEvent(cmd *common.Command, debug string) error {
	args := &DebugEvent{
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		Funds:          cmd.Amount,
		DebugMessage:   debug,
		TransactionNum: cmd.TransactionID,
	}
	return l.Call(debugEventMethod, args, nil)
}

func (l *logger) DumpLogUser(userId string) (*[]byte, error) {
	var data []byte
	err := l.Call(dumpLogMethod, DumpLogArgs{UserId: userId}, &data)
	return &data, err
}

func (l *logger) DumpLog() (*[]byte, error) {
	var data []byte
	err := l.Call(dumpLogMethod, DumpLogArgs{UserId: "admin"}, &data)
	return &data, err
}

func (l *logger) Close() error {
	return l.client.Close()
}

func GetLogger(server string) Logger {
	for {
		client, err := rpc.Dial("tcp", common.CFG.AuditServer.Url)
		if err != nil {
			log.Println(err)
			continue
		}
		return &logger{client, server}
	}
}

type LoggerRPC struct {
	db *MongoDB
}

func (l *LoggerRPC) readLog(userid string) ([]byte, error) {
	data := []byte("<log>\n")
	logs, err := l.db.Logs.GetLogs(userid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, val := range logs {
		toWrite := append(val.Xml, byte('\n'))
		data = append(data, toWrite...)
	}
	data = append(data, []byte("\n</log>")...)
	return data, nil
}

func (l *LoggerRPC) writeLog(e interface{}, userid string) error {
	data, err := xml.MarshalIndent(e, "  ", "    ")
	if err != nil {
		return err
	}
	eLog := &common.EventLog{UserId: userid, Xml: data}
	l.db.Logs.LogEvent(eLog)
	return nil
}

func (l *LoggerRPC) UserCommand(cmd *UserCommand, result *string) error {
	return l.writeLog(cmd, cmd.Username)
}

func (l *LoggerRPC) QuoteServer(qs *QuoteServer, result *string) error {
	return l.writeLog(qs, qs.Username)
}

func (l *LoggerRPC) AccountTransaction(txn *AccountTransaction, result *string) error {
	return l.writeLog(txn, txn.Username)
}

func (l *LoggerRPC) SystemEvent(e *SystemEvent, result *string) error {
	return l.writeLog(e, e.Username)
}

func (l *LoggerRPC) ErrorEvent(e *ErrorEvent, result *string) error {
	return l.writeLog(e, e.Username)
}

func (l *LoggerRPC) DebugEvent(e *DebugEvent, result *string) error {
	return l.writeLog(e, e.Username)
}

func (l *LoggerRPC) DumpLog(args *DumpLogArgs, result *[]byte) error {
	var err error
	*result, err = l.readLog(args.UserId)
	return err
}

func GetLoggerRPC() (*LoggerRPC, *MongoDB) {
	log.Println("Attempting to initiate RPC")
	db := GetMongoDatabase()
	return &LoggerRPC{db}, db
}
