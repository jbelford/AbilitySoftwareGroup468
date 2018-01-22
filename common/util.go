package common

import (
	"encoding/json"
	"os"
)

func GetConfig() (Config, error) {
	var config Config
	file, err := os.Open("../config/config.json")
	if err != nil {
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}
