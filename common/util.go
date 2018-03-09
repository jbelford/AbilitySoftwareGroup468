package common

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var CFG Config

func init() {
	var configPath = "./config/config.local.json"
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

func EncodeData(obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		log.Println("Failed encoding data: " + err.Error()) // Should not happen
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeData(data []byte, obj interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(obj)
	if err != nil {
		log.Println("Failed decoding data: " + err.Error()) // Should also not happen
	}
	return err
}
