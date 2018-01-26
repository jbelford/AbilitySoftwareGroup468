package logging

type Logger interface {
	UserCommand(args *Args, result *string) error
	QuoteServer(args *Args, result *string) error
	AccountTransaction(args *Args, result *string) error
	SystemEvent(args *Args, result *string) error
	ErrorEvent(args *Args, result *string) error
	DebugEvent(args *Args, result *string) error
}

type logger struct {
}
