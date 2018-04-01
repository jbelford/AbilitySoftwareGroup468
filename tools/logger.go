package tools

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"
	"strings"

	"github.com/valyala/gorpc"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

const (
	LoggerServiceName = "LoggerRPC"

	userCommandMethod = "UserCommand"
	quoteServerMethod = "QuoteServer"
	accountTxnMethod  = "AccountTransaction"
	systemEventMethod = "SystemEvent"
	errorEventMethod  = "ErrorEvent"
	debugEventMethod  = "DebugEvent"
	dumpLogMethod     = "DumpLog"
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
	Close()
}

type logger struct {
	client   *gorpc.Client
	dispatch *gorpc.DispatcherClient
	server   string
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
	_, err := l.dispatch.Call(userCommandMethod, args)
	return err
}

func (l *logger) QuoteServer(quote *common.QuoteData, tid int64) error {
	// Convert integer price to string without using floating point arithmetic
	price := fmt.Sprintf("%03d", quote.Quote) // Pads zeros if length less than 3
	mid := len(price) - 2
	price = price[:mid] + "." + price[mid:]
	cryptokey := strings.TrimSpace(quote.Cryptokey)

	args := &QuoteServer{
		TransactionNum:  tid,
		Timestamp:       uint64(time.Now().Unix() * 1000),
		Server:          l.server,
		Username:        quote.UserId,
		Price:           price,
		StockSymbol:     quote.Symbol,
		QuoteServerTime: quote.Timestamp,
		Cryptokey:       cryptokey,
	}
	_, err := l.dispatch.Call(quoteServerMethod, args)
	return err
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
	_, err := l.dispatch.Call(accountTxnMethod, args)
	return err
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
	_, err := l.dispatch.Call(systemEventMethod, args)
	return err
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
	_, err := l.dispatch.Call(errorEventMethod, args)
	return err
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
	_, err := l.dispatch.Call(debugEventMethod, args)
	return err
}

func (l *logger) DumpLogUser(userId string) (*[]byte, error) {
	data, err := l.dispatch.Call(dumpLogMethod, &DumpLogArgs{UserId: userId})
	if err != nil {
		return nil, err
	}
	bytes, _ := data.([]byte)
	return &bytes, nil
}

func (l *logger) DumpLog() (*[]byte, error) {
	data, err := l.dispatch.Call(dumpLogMethod, &DumpLogArgs{UserId: "admin"})
	if err != nil {
		return nil, err
	}
	bytes, _ := data.([]byte)
	return &bytes, nil
}

func (l *logger) Close() {
	l.client.Stop()
}

func GetLogger(server string) Logger {
	gorpc.RegisterType(&UserCommand{})
	gorpc.RegisterType(&QuoteServer{})
	gorpc.RegisterType(&AccountTransaction{})
	gorpc.RegisterType(&SystemEvent{})
	gorpc.RegisterType(&ErrorEvent{})
	gorpc.RegisterType(&DebugEvent{})
	gorpc.RegisterType(&DumpLogArgs{})

	client := gorpc.NewTCPClient(common.CFG.AuditServer.Url)
	dispatcher := gorpc.NewDispatcher()
	dispatcher.AddService(LoggerServiceName, &LoggerRPC{})
	dispatchClient := dispatcher.NewServiceClient(LoggerServiceName, client)
	client.Start()
	return &logger{client, dispatchClient, server}
}

type LoggerRPC struct {
	session *MongoSession
	work    chan *common.EventLog
	flush   chan bool
}

func (l *LoggerRPC) readLog(userid string) ([]byte, error) {
	l.flush <- true
	// While flushing may as well set socket connection ready
	db := l.session.GetUniqueInstance()
	defer db.Close()

	<-l.flush

	data := []byte("<log>\n")
	logs, err := db.Logs.GetLogs(userid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, val := range logs {
		toWrite := append(val.Xml, byte('\n'))
		data = append(data, toWrite...)
	}
	data = append(data, []byte("</log>")...)

	if f, err := os.OpenFile("log.xml", os.O_WRONLY|os.O_CREATE, 0777); err == nil {
		defer f.Close()
		f.Write(data)
	}

	return data, nil
}

func (l *LoggerRPC) writeLog(e interface{}, userid string) error {
	data, err := xml.MarshalIndent(e, "  ", "    ")
	if err != nil {
		return err
	}
	l.work <- &common.EventLog{UserId: userid, Xml: data}
	return nil
}

func (l *LoggerRPC) UserCommand(cmd *UserCommand) error {
	return l.writeLog(cmd, cmd.Username)
}

func (l *LoggerRPC) QuoteServer(qs *QuoteServer) error {
	return l.writeLog(qs, qs.Username)
}

func (l *LoggerRPC) AccountTransaction(txn *AccountTransaction) error {
	return l.writeLog(txn, txn.Username)
}

func (l *LoggerRPC) SystemEvent(e *SystemEvent) error {
	return l.writeLog(e, e.Username)
}

func (l *LoggerRPC) ErrorEvent(e *ErrorEvent) error {
	return l.writeLog(e, e.Username)
}

func (l *LoggerRPC) DebugEvent(e *DebugEvent) error {
	return l.writeLog(e, e.Username)
}

func (l *LoggerRPC) DumpLog(args *DumpLogArgs) ([]byte, error) {
	return l.readLog(args.UserId)
}

func GetLoggerRPC(session *MongoSession) *LoggerRPC {
	log.Println("Attempting to initiate RPC")
	l := &LoggerRPC{session, nil, nil}
	l.initBulkProcessing()
	return l
}

func (l *LoggerRPC) initBulkProcessing() {
	limit := 100000
	l.work = make(chan *common.EventLog, limit)
	l.flush = make(chan bool)
	go func() {
		db := l.session.GetSharedInstance()
		defer db.Close()
		for {
			eventLogs, flushed := l.wait(limit)
			if len(eventLogs) > 0 {
				db.Logs.LogEvents(eventLogs)
			}
			if flushed {
				l.flush <- true
			}
		}
	}()
}

func (l *LoggerRPC) wait(limit int) ([]*common.EventLog, bool) {
	eventLogs := []*common.EventLog{}
	for i := 0; i < limit; i++ {
		select {
		case val := <-l.work:
			log.Printf("%s\n",val)
			eventLogs = append(eventLogs, val)
			log.Printf("Properly appended!")
		case <-l.flush:
			return eventLogs, true
		}
	}
	return eventLogs, false
}
