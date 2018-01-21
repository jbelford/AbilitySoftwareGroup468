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
	command_type, err := strconv.ParseInt(parsed[0], 10, 0)

	defer command.commands[int(command_type)](parsed)

}
