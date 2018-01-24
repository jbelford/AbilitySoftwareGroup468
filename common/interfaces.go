package common

import (
	"net"
	"time"
)

type Server interface {
	Start()
}

type Response struct {
	Success bool
	Message string `json:",omitempty"`
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
	Quote     int
	Symbol    string
	UserId    string
	Timestamp uint64
	Cryptokey string
}

// Mock struct of net.Conn
type MockConn struct {
	data string
}

func (c *MockConn) Write(b []byte) (n int, e error) {
	return len(b), nil
}

func (c *MockConn) Read(b []byte) (n int, e error) {
	b = []byte(c.data)
	return len(b), nil
}

func (c *MockConn) Close() error                       { return nil }
func (c *MockConn) LocalAddr() net.Addr                { return nil }
func (c *MockConn) RemoteAddr() net.Addr               { return nil }
func (c *MockConn) SetDeadline(t time.Time) error      { return nil }
func (c *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func main() {

}
