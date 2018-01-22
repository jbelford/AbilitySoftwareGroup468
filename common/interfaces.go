package common

type Server interface {
	Start()
}

type Config struct {
	database    DatabaseConfig
	quoteserver QuoteServConfig
}

type DatabaseConfig struct {
	url  string
	name string
}

type QuoteServConfig struct {
	address string
}

type QuoteData struct {
	Quote     int
	Symbol    string
	UserId    string
	Timestamp uint64
	Cryptokey string
}

func main() {

}
