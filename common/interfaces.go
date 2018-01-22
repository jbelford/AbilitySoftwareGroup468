package common

import (
	"net"
	"time"
)

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
	mock    bool
}

type QuoteData struct {
	Quote     int
	Symbol    string
	UserId    string
	Timestamp uint64
	Cryptokey string
}

// Can be used to mock net.Conn
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
