package common

import (
	"encoding/json"
	"log"
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
	C_type        int
	TransactionID int
	UserId        string
	Amount        int64     `json:",omitempty"`
	StockSymbol   string    `json:",omitempty"`
	FileName      string    `json:",omitempty"`
	Timestamp     time.Time `json:",omitempty"`
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
}
