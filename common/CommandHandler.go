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
	c_type      int
	userid      string
	amount      int
	stockSymbol string
	filename    string
}

func (cmd *Command) commandObjToString() string {
	return str(cmd.c_type) + "," + cmd.userid + "," + str(cmd.amount) + "," + cmd.stockSymbol + "," + cmd.filename
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
