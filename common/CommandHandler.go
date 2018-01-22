package common

import (
	"log"
	"strconv"
	"strings"
)

type CommandHandler struct {
	commands map[int]func(Command)
}

type Command struct {
	C_type      int
	UserId      string
	Amount      int
	StockSymbol string
	FileName    string
}

func (cmd *Command) commandObjToString() string {
	return string(cmd.C_type) + "," + cmd.UserId + "," + string(cmd.Amount) + "," + cmd.StockSymbol + "," + cmd.FileName
}

func commandConstructor(com string) Command {
	com := com.split(",")
	com[0] = commandToInt(com[0])
	p := new(Command)
	p.C_type = com[0]
	switch p.C_type {
	case common.ADD:
		p.UserId = com[1]
		p.Amount = com[2]
	case common.QUOTE:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case common.BUY:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		p.Amount = com[3]
	case common.COMMIT_BUY:
		p.UserId = com[1]
	case common.CANCEL_BUY:
		p.UserId = com[1]
	case common.SELL:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		p.Amount = com[3]
	case common.COMMIT_SELL:
		p.UserId = com[1]
	case common.CANCEL_SELL:
		p.UserId = com[1]
	case common.SET_BUY_AMOUNT:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		p.Amount = com[3]
	case common.CANCEL_SET_BUY:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case common.SET_BUY_TRIGGER:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		p.Amount = com[3]
	case common.SET_SELL_AMOUNT:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		p.Amount = com[3]
	case common.SET_SELL_TRIGGER:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		p.Amount = com[3]
	case common.CANCEL_SET_SELL:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case common.DUMPLOG:
		if len(com) == 3{
			p.UserId = com[1]
			p.FileName = com[2]
		}
		else{
			p.C_type = common.ADMIN_DUMPLOG
			p.FileName = com[1]
		}
	case common.DISPLAY_SUMMARY:
		p.UserId = com[1]
	}
	return p
}

func (command *CommandHandler) On(command_name int, function_to_call func(args Command)) {
	command.commands[command_name] = function_to_call
}

func (command *CommandHandler) parse(commandStr string) {

	log.Println("Received!:", string(commandStr))
	parsed := strings.Split(commandStr, ",")
	if len(parsed) < 5 {
		log.Println("Error! Not enough arguments!!!")
		return
	}

	command_type, _ := strconv.ParseInt(parsed[0], 10, 0)
	userid := parsed[1]
	amount, _ := strconv.ParseFloat(parsed[2], 10)
	stockSymbol := parsed[3]
	filename := parsed[4]

	command_obj := Command{int(command_type), userid, int(amount), stockSymbol, filename}

	defer command.commands[int(command_type)](command_obj)
}
