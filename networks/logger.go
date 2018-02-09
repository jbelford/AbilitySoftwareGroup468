package networks

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
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
	Close() error
}

type logger struct {
	client *rpc.Client
	server string
}

func (l *logger) connect() {
	var err error
	for {
		l.client, err = rpc.Dial("tcp", common.CFG.AuditServer.Url)
		if err != nil {
			log.Println("FAILED TO CONNECT TO AUDITSERVER")
			continue
		}
		break
	}
}

func (l *logger) Call(method string, args interface{}, result interface{}) (err error) {
	for {
		err = l.client.Call(method, args, result)
		if err != nil {
			l.connect()
			continue
		}
		return err
	}
}

func (l *logger) UserCommand(cmd *common.Command) error {
	args := &Args{
		TransactionNum: cmd.TransactionID,
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
	}
	return l.Call(userCommandMethod, args, nil)
}

func (l *logger) QuoteServer(quote *common.QuoteData, tid int64) error {
	args := &Args{
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
	args := &Args{
		TransactionNum: tid,
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Action:         action,
		Username:       userId,
		Funds:          funds,
	}
	return l.Call(accountTransactionMethod, args, nil)
}

func (l *logger) SystemEvent(cmd *common.Command) error {
	args := &Args{
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
		TransactionNum: cmd.TransactionID,
	}
	return l.Call(systemEventMethod, args, nil)
}

func (l *logger) ErrorEvent(cmd *common.Command, e string) error {
	args := &Args{
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
		TransactionNum: cmd.TransactionID,
		ErrorMessage:   e,
	}
	return l.Call(errorEventMethod, args, nil)
}

func (l *logger) DebugEvent(cmd *common.Command, debug string) error {
	args := &Args{
		Timestamp:      uint64(time.Now().Unix() * 1000),
		Server:         l.server,
		Command:        common.Commands[cmd.C_type],
		Username:       cmd.UserId,
		StockSymbol:    cmd.StockSymbol,
		FileName:       cmd.FileName,
		Funds:          cmd.Amount,
		DebugMessage:   debug,
		TransactionNum: cmd.TransactionID,
	}
	return l.Call(debugEventMethod, args, nil)
}

func (l *logger) DumpLogUser(userId string) (*[]byte, error) {
	var data []byte
	err := l.Call(dumpLogMethod, Args{FileName: userId + ".xml"}, &data)
	return &data, err
}

func (l *logger) DumpLog() (*[]byte, error) {
	var data []byte
	err := l.Call(dumpLogMethod, Args{}, &data)
	return &data, err
}

func (l *logger) Close() error {
	return l.client.Close()
}

func GetLogger(server string) Logger {
	l := &logger{server: server}
	l.connect()
	return l
}

type LoggerRPC struct {
	writer *os.File
}

func (l *LoggerRPC) writeLogs(log interface{}, userFilename string) error {
	flag := os.O_APPEND | os.O_WRONLY
	if _, err := os.Stat("./logs/" + userFilename); os.IsNotExist(err) {
		flag |= os.O_CREATE
	}
	uWriter, err := os.OpenFile("./logs/"+userFilename, flag, 0777)
	if err != nil {
		return err
	}
	defer uWriter.Close()
	toWrite, err := xml.MarshalIndent(log, "  ", "    ")
	if err == nil {
		toWrite := append(toWrite, '\n')
		l.writer.Write(toWrite)
		uWriter.Write(toWrite)
	}
	return err
}

func (l *LoggerRPC) readLog(filename string) []byte {
	data := []byte("<log>\n")
	read, err := ioutil.ReadFile("./logs/" + filename)
	if err == nil {
		data = append(data, read...)
	}
	data = append(data, []byte("\n</log>")...)
	return data
}

func (l *LoggerRPC) UserCommand(args *Args, result *string) error {
	val := &UserCommand{
		Command:        args.Command,
		Server:         args.Server,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	log.Println("USER_COMMAND", val)
	return l.writeLogs(val, args.Username+".xml")
}

func (l *LoggerRPC) QuoteServer(args *Args, result *string) error {
	val := &QuoteServer{
		Cryptokey:       args.Cryptokey,
		Server:          args.Server,
		Price:           args.Price,
		QuoteServerTime: args.QuoteServerTime,
		StockSymbol:     args.StockSymbol,
		Timestamp:       args.Timestamp,
		TransactionNum:  args.TransactionNum,
		Username:        args.Username,
	}
	log.Println("QUOTE_SERVER", val)
	return l.writeLogs(val, args.Username+".xml")
}

func (l *LoggerRPC) AccountTransaction(args *Args, result *string) error {
	val := &AccountTransaction{
		Action:         args.Action,
		Server:         args.Server,
		Funds:          args.Funds,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	log.Println("ACCOUNT_TRANSACTION", val)
	return l.writeLogs(val, args.Username+".xml")
}

func (l *LoggerRPC) SystemEvent(args *Args, result *string) error {
	val := &SystemEvent{
		Command:        args.Command,
		Server:         args.Server,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	log.Println("SYSTEM_EVENT", val)
	return l.writeLogs(val, args.Username+".xml")
}

func (l *LoggerRPC) ErrorEvent(args *Args, result *string) error {
	val := &ErrorEvent{
		Command:        args.Command,
		Server:         args.Server,
		ErrorMessage:   args.ErrorMessage,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	log.Println("ERROR_EVENT", val)
	return l.writeLogs(val, args.Username+".xml")
}

func (l *LoggerRPC) DebugEvent(args *Args, result *string) error {
	val := &DebugEvent{
		Command:        args.Command,
		Server:         args.Server,
		DebugMessage:   args.DebugMessage,
		Filename:       args.FileName,
		Funds:          args.Funds,
		StockSymbol:    args.StockSymbol,
		Timestamp:      args.Timestamp,
		TransactionNum: args.TransactionNum,
		Username:       args.Username,
	}
	log.Println("DEBUG_EVENT", val)
	return l.writeLogs(val, args.Username+".xml")
}

func (l *LoggerRPC) DumpLog(args *Args, result *[]byte) error {
	filename := "./logs/tmp.xml"
	if args.FileName == "" {
		filename = args.FileName
	}
	*result = l.readLog(filename)
	return nil
}

func GetLoggerRPC() (*LoggerRPC, *os.File) {
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		if err = os.Mkdir("logs", 0777); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("log folder made")

	flag := os.O_APPEND | os.O_WRONLY
	if _, err := os.Stat("./logs/tmp.xml"); os.IsNotExist(err) {
		flag |= os.O_CREATE
	}
	writer, err := os.OpenFile("./logs/tmp.xml", flag, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return &LoggerRPC{writer}, writer
}
