package common

import (
	"log"
	"strconv"
	"strings"
)

type CommandHandler struct {
	commands map[int]func(Command)
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

	command_type, err := strconv.ParseInt(parsed[0], 10, 0)
	userid := parsed[1]
  amount, err2 := strconv.ParseFloat(parsed[2], 10)
  stockSymbol := parsed[3]
  filename := parsed[4]

	command_obj = Command{command_type, userid, amount, stockSymbol, filename}

	defer command.commands[int(command_type)](command_obj)

}
