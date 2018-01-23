package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var configPath = "../config/config.local.json"

func init() {
	if len(os.Args) > 2 && os.Args[2] == "--prod" {
		configPath = "../config/config.prod.json"
	}
}

func GetConfig() (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}
