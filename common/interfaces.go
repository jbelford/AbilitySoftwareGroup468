package common

import (
	"time"
)

type Server interface {
	Start()
}

type Response struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message,omitempty"`
	Stock        string         `json:"stock,omitempty"`
	Quote        int64          `json:"quote,omitempty"`
	ReqAmount    int64          `json:"amount_requested,omitempty"`
	RealAmount   int64          `json:"real_amount,omitempty"`
	Shares       int            `json:"shares,omitempty"`
	Expiration   int64          `json:"expiration,omitempty"`
	Paid         int64          `json:"paid,omitempty"`
	Received     int64          `json:"received,omitempty"`
	SharesAfford int            `json:"shares_afford,omitempty"`
	AffordAmount int64          `json:"afford_amount,omitempty"`
	Status       *UserInfo      `json:"status,omitempty"`
	Transactions *[]Transaction `json:"transactions,omitempty"`
	Triggers     *[]Trigger     `json:"triggers,omitempty"`
	File         *[]byte        `json:"file,omitempty"`
}

type Command struct {
	C_type        int
	TransactionID int64
	UserId        string
	Amount        int64     `json:",omitempty"`
	StockSymbol   string    `json:",omitempty"`
	FileName      string    `json:",omitempty"`
	Timestamp     time.Time `json:",omitempty"`
}

type UserInfo struct {
	Balance  int64 `json:"balance"`
	Reserved int64 `json:"reserved"`
	Stock    map[string]struct {
		Real     int `json:"real"`
		Reserved int `json:"reserved"`
	} `json:"stock,omitempty"`
}

type Transactions struct {
	UserId string `bson:"_id" json:"userId"`
	Logged []Transaction
}

type Transaction struct {
	Type      string `json:"type"`
	Triggered bool   `json:"triggered"`
	Stock     string `json:"stock"`
	Amount    int64  `json:"amount"`
	Shares    int    `json:"shares"`
	Timestamp uint64 `json:"timestamp"`
}

type PendingTxn struct {
	UserId   string
	Type     string
	Stock    string
	Reserved int64
	Price    int64
	Shares   int
	Expiry   time.Time
}

type Config struct {
	Database struct {
		Url  string `json:"url"`
		LUrl string `json:"lurl"`
		Name string `json:"name"`
	} `json:"database"`
	Quoteserver struct {
		Address string `json:"address"`
		Mock    bool   `json:"mock"`
	} `json:"quoteserver"`
	TxnServer struct {
		Url  string `json:"url"`
		LUrl string `json:"lurl"`
	} `json:"transactionserver"`
	WebServer struct {
		Url  string `json:"url"`
		LUrl string `json:"lurl"`
	} `json:"webserver"`
	AuditServer struct {
		Url  string `json:"url"`
		LUrl string `json:"lurl"`
	} `json:"auditserver"`
	Logging struct {
		Db bool `json:"db"`
	} `json:"logging"`
}

type QuoteData struct {
	Quote     int64
	Symbol    string
	UserId    string
	Timestamp uint64
	Cryptokey string
}

type User struct {
	UserId   string `bson:"_id" json:"userId"`
	Balance  int64  `json:"balance"`
	Reserved int64  `json:"reserved"`
	Stock    map[string]struct {
		Real     int `json:"real"`
		Reserved int `json:"reserved"`
	} `json:"stock"`
}

type Trigger struct {
	UserId        string `bson:"userId" json:"userId"`
	Stock         string `json:"stock"`
	TransactionID int64  `bson:"transactionId" json:"transactionId"`
	Type          string `json:"type"`
	Shares        int    `json:"shares"`
	Amount        int64  `json:"amount"`
	When          int64  `json:"when"`
}

type EventLog struct {
	UserId string `bson:"userId" json:"userId"`
	Xml    []byte `json:"xml"`
}
