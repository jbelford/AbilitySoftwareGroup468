package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var CFG Config

func init() {
	var configPath = "./config/config.local.json"
	if len(os.Args) > 2 && os.Args[2] == "--prod" {
		configPath = "./config/config.prod.json"
	}
	var config Config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &config)
	CFG = config
}
