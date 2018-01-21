
package main
type Server interface {
	Start()
}

type Config struct {
	database DatabaseConfig
}

type DatabaseConfig struct {
	url  string
	name string
}

func main() {

}
