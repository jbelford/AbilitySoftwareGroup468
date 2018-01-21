package common

import (
	"log"
	"strings"
)


type CommandHandler struct{
	commands map[int]func(args Command)
}

func (command *CommandHandler) On(command_name int, function_to_call func(args Command)) {
	command.commands[command_name] = function_to_call
}

func (command *CommandHandler) parse(command string) {
	command_hex := CommandHexReplacement{}

	log.Println("Received!:", string(command))
	parsed = strings.Split(command, ",")
	command_type = command_hex.toCommandHex(parsed[0])

	defer command.commands[command_type](parsed)

}
