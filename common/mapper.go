package common

type CommandHexReplacement int

const (
	ADD int = 0 + iota
	QUOTE
	BUY
	COMMIT_BUY
	CANCEL_BUY
	SELL
	COMMIT_SELL
	CANCEL_SELL
	SET_BUY_AMOUNT
	CANCEL_SET_BUY
	SET_BUY_TRIGGER
	SET_SELL_AMOUNT
	SET_SELL_TRIGGER
	CANCEL_SET_SELL
	DUMPLOG
	ADMIN_DUMPLOG
	DISPLAY_SUMMARY
)

var Commands = [...]string{
	"ADD",
	"QUOTE",
	"BUY",
	"COMMIT_BUY",
	"CANCEL_BUY",
	"SELL",
	"COMMIT_SELL",
	"CANCEL_SELL",
	"SET_BUY_AMOUNT",
	"CANCEL_SET_BUY",
	"SET_BUY_TRIGGER",
	"SET_SELL_AMOUNT",
	"SET_SELL_TRIGGER",
	"CANCEL_SET_SELL",
	"DUMPLOG",
	"ADMIN_DUMPLOG",
	"DISPLAY_SUMMARY",
}

func (chr CommandHexReplacement) commandToString() string {
	return Commands[chr-1]
}

func CommandToInt(s string) int {
	for k, v := range Commands {
		if s == v {
			return k
		}
	}
	return -1
}
