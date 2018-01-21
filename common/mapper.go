package common

type CommandHexReplacement int

const (
	ADD CommandHexReplacement = 1 + iota
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

// String() function will return the english name
// that we want out constant Day be recognized as
func (chr CommandHexReplacement) String() string {
	return Commands[chr-1]
}
