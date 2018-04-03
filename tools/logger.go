package tools

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

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

	args := &QuoteServer{
		TransactionNum:  tid,
		Timestamp:       uint64(time.Now().Unix() * 1000),
		Server:          l.server,
		Username:        quote.UserId,
		Price:           price,
		StockSymbol:     quote.Symbol,
		QuoteServerTime: quote.Timestamp,
		Cryptokey:       quote.Cryptokey,
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
	<-l.flush

	data := []byte("<log>\n")

	var err error
	var logs []common.EventLog
	if common.CFG.Logging.Db {
		db := l.session.GetUniqueInstance()
		defer db.Close()

		logs, err = db.Logs.GetLogs(userid)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		data, err := ioutil.ReadFile("log.json")
		if err == nil {
			var eventLogs []common.EventLog
			json.Unmarshal(data, &eventLogs)
			for _, l := range eventLogs {
				if userid == "admin" || userid == l.UserId {
					logs = append(logs, l)
				}
			}
		}
	}
	for _, l := range logs {
		toWrite := append(l.Xml, byte('\n'))
		data = append(data, toWrite...)
	}
	data = append(data, []byte("</log>")...)

	os.Remove("log.xml")
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
		for {
			eventLogs, flushed := l.wait(limit)
			if len(eventLogs) > 0 {
				l.bulkWrite(eventLogs)
			}
			if flushed {
				l.flush <- true
			}
		}
	}()
}

func (l *LoggerRPC) bulkWrite(eventLogs []*common.EventLog) {
	if common.CFG.Logging.Db {
		db := l.session.GetSharedInstance()
		defer db.Close()
		db.Logs.LogEvents(eventLogs)
	} else {
		jsonLogs := make([]*common.EventLog, 0)
		data, err := ioutil.ReadFile("log.json")
		if err == nil {
			err = json.Unmarshal(data, &jsonLogs)
			if err != nil {
				log.Println("Failed to unmarshal json data: " + err.Error())
			}
		}
		jsonLogs = append(jsonLogs, eventLogs...)

		data, err = json.Marshal(jsonLogs)
		if err != nil {
			log.Println("Failed to marshal json data: " + err.Error())
			return
		}
		os.Remove("log.json")
		f, err := os.OpenFile("log.json", os.O_APPEND|os.O_CREATE, 0777)
		if err != nil {
			log.Println("Error writing to file: " + err.Error())
			return
		}
		defer f.Close()
		_, err = f.Write(data)
		if err != nil {
			log.Println("Error writing data: " + err.Error())
		}
	}
}

func (l *LoggerRPC) wait(limit int) ([]*common.EventLog, bool) {
	eventLogs := []*common.EventLog{}
	for i := 0; i < limit; i++ {
		select {
		case val := <-l.work:
			eventLogs = append(eventLogs, val)
		case <-l.flush:
			return eventLogs, true
		}
	}
	return eventLogs, false
}
