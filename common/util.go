package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var CFG Config

func init() {
	var configPath = "./config/config.prod.json"
	if len(os.Args) > 2 && os.Args[2] == "--prod" {
		configPath = "./config/config.prod.json"
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &CFG)
	if err != nil {
		panic(err)
	}
}
