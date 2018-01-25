package common

import (
	"time"
)

type Server interface {
	Start()
}

type Response struct {
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	Stock        string `json:"stock,omitempty"`
	Quote        int64  `json:"quote,omitempty"`
	ReqAmount    int64  `json:"amount_requested,omitempty"`
	RealAmount   int64  `json:"real_amount,omitempty"`
	Shares       int    `json:"shares,omitempty"`
	Expiration   int64  `json:"expiration,omitempty"`
	Paid         int64  `json:"paid,omitempty"`
	Received     int64  `json:"received,omitempty"`
	SharesAfford int    `json:"shares_afford,omitempty"`
	AffordAmount int64  `json:"afford_amount,omitempty"`
}

type PendingTxn struct {
	Type   string
	Stock  string
	Price  int64
	Shares int
	Expiry time.Time
}

type Config struct {
	Database    DatabaseConfig  `json:"database"`
	Quoteserver QuoteServConfig `json:"quoteserver"`
}

type DatabaseConfig struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type QuoteServConfig struct {
	Address string `json:"address"`
	Mock    bool   `json:"mock"`
}

type QuoteData struct {
	Quote     int64
	Symbol    string
	UserId    string
	Timestamp uint64
	Cryptokey string
}

type User struct {
	UserId   string         `json:"_id"`
	Balance  int64          `json:"balance"`
	Stock    map[string]int `json:"stock"`
	Triggers []string       `json:"triggers"`
}

func main() {

}
