package common

import (
	"encoding/json"
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
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}
