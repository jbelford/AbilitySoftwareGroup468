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
	case 0: //add
		p.Amount = com[1]
	case 1: //quote
		p.StockSymbol = com[1]
	case 2:

	case 3:
	case 4:
	case 5:
	case 6:
	case 7:
	case 8:
	case 9:
	case 10:
	case 11:
	case 12:
	case 13:
	case 14:
	case 15:
	case 16:

	}
	p.C_type = 0
	p.UserId = ""
	p.Amount = 0
	p.StockSymbol = ""
	p.FileName = ""
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
