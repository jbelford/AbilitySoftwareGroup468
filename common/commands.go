package common

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"
)

type CommandHandler struct {
	commands map[int]func(*Command) *Response
}

func NewCommandHandler() *CommandHandler {
	var h CommandHandler
	h.commands = make(map[int]func(*Command) *Response)
	return &h
}

type Command struct {
	C_type      int
	UserId      string
	Amount      int64     `json:",omitempty"`
	StockSymbol string    `json:",omitempty"`
	FileName    string    `json:",omitempty"`
	Timestamp   time.Time `json:",omitempty"`
}

func (cmd *Command) CommandObjToString() string {
	return string(cmd.C_type) + "," + cmd.UserId + "," + string(cmd.Amount) + "," + cmd.StockSymbol + "," + cmd.FileName
}

func CommandConstructor(cmd string) Command {
	trim_cmd := strings.TrimSpace(cmd)
	com := strings.Split(trim_cmd, ",")
	log.Print(com)
	p := new(Command)
	p.C_type = CommandToInt(com[0])
	switch p.C_type {
	case ADD:
		p.UserId = com[1]
		temp_amount, err := strconv.ParseInt(com[2], 10, 32)
		log.Print(err)
		p.Amount = temp_amount
	case QUOTE:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case BUY:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseInt(com[3], 10, 32)
		p.Amount = temp_amount
	case COMMIT_BUY:
		p.UserId = com[1]
	case CANCEL_BUY:
		p.UserId = com[1]
	case SELL:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseInt(com[3], 10, 32)
		p.Amount = temp_amount
	case COMMIT_SELL:
		p.UserId = com[1]
	case CANCEL_SELL:
		p.UserId = com[1]
	case SET_BUY_AMOUNT:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseInt(com[3], 10, 32)
		p.Amount = temp_amount
	case CANCEL_SET_BUY:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case SET_BUY_TRIGGER:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseInt(com[3], 10, 32)
		p.Amount = temp_amount
	case SET_SELL_AMOUNT:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseInt(com[3], 10, 32)
		p.Amount = temp_amount
	case SET_SELL_TRIGGER:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseInt(com[3], 10, 32)
		p.Amount = temp_amount
	case CANCEL_SET_SELL:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case DUMPLOG:
		if len(com) == 3 {
			p.UserId = com[1]
			p.FileName = com[2]
		} else {
			p.C_type = ADMIN_DUMPLOG
			p.FileName = com[1]
		}
	case DISPLAY_SUMMARY:
		p.UserId = com[1]
	}
	return *p
}

func (command *CommandHandler) On(command_name int, function_to_call func(args *Command) *Response) {
	command.commands[command_name] = function_to_call
}

func (command *CommandHandler) Parse(commandStr string) (*Response, error) {
	log.Println("Received!:", string(commandStr))

	var cmd *Command
	err := json.Unmarshal([]byte(commandStr), &cmd)
	if err != nil {
		return nil, err
	}

	return command.commands[cmd.C_type](cmd), nil

	// parsed := strings.Split(commandStr, ",")
	// if len(parsed) < 5 {
	// 	log.Println("Error! Not enough arguments!!!")
	// 	return
	// }

	// command_type, _ := strconv.ParseInt(parsed[0], 10, 0)
	// userid := parsed[1]
	// amount, _ := strconv.ParseInt(parsed[2], 10)
	// stockSymbol := parsed[3]
	// filename := parsed[4]

	// command_obj := Command{int(command_type), userid, amount), stockSymbol, filename, time.Now()}

}
