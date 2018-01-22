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

func CommandConstructor(cmd string) Command {
	trim_cmd := strings.TrimSpace(cmd)
	com := strings.Split(trim_cmd, ",")
	log.Print(com)
	p := new(Command)
	p.C_type = CommandToInt(com[0])
	switch p.C_type {
	case ADD:
		p.UserId = com[1]
		temp_amount, err := strconv.ParseFloat(com[2], 32)
		log.Print(err)
		p.Amount = int(temp_amount * 100)
	case QUOTE:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case BUY:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseFloat(com[3], 32)
		p.Amount = int(temp_amount * 100)
	case COMMIT_BUY:
		p.UserId = com[1]
	case CANCEL_BUY:
		p.UserId = com[1]
	case SELL:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseFloat(com[3], 32)
		p.Amount = int(temp_amount * 100)
	case COMMIT_SELL:
		p.UserId = com[1]
	case CANCEL_SELL:
		p.UserId = com[1]
	case SET_BUY_AMOUNT:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseFloat(com[3], 32)
		p.Amount = int(temp_amount * 100)
	case CANCEL_SET_BUY:
		p.UserId = com[1]
		p.StockSymbol = com[2]
	case SET_BUY_TRIGGER:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseFloat(com[3], 32)
		p.Amount = int(temp_amount * 100)
	case SET_SELL_AMOUNT:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseFloat(com[3], 32)
		p.Amount = int(temp_amount * 100)
	case SET_SELL_TRIGGER:
		p.UserId = com[1]
		p.StockSymbol = com[2]
		temp_amount, _ := strconv.ParseFloat(com[3], 32)
		p.Amount = int(temp_amount * 100)
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

func (command *CommandHandler) On(command_name int, function_to_call func(args Command)) {
	command.commands[command_name] = function_to_call
}

func (command *CommandHandler) Parse(commandStr string) {

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
