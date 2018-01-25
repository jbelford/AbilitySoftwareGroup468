package common

type Server interface {
	Start()
}

type Response struct {
	Success bool
	Message string `json:",omitempty"`
	Stock   string `json:",omitempty"`
	Quote   int64  `json:",omitempty"`
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

func main() {

}
