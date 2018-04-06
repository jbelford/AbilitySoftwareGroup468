package common

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
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
func ClusterDialTCP(address string, data []byte, number int) ([]byte, error) {
	ch := make(chan []byte, number)
	for i := 0; i < number; i++ {
		go dialTCP(address, data, ch)
	}
	for {
		data := <-ch
		if data != nil {
			return data, nil
		}
	}
	return nil, errors.New("ClusterDialTCP: No connections succeeded")
}

func dialTCP(address string, data []byte, ch chan []byte) {
	tcpConn, err := net.Dial("tcp", address)
	if err != nil {
		ch <- nil
		return
	}
	defer tcpConn.Close()
	if data != nil {
		_, err = tcpConn.Write(data)
		if err != nil {
			log.Printf("dialTCP: Failed to write data to connection '%s' - %s", address, err.Error())
			ch <- nil
			return
		}
	}
	respData, err := bufio.NewReader(tcpConn).ReadBytes(byte('\n'))
	if err != nil {
		log.Printf("dialTCP: Failed to read data from connection '%s' - %s", address, err.Error())
		ch <- nil
		return
	}
	ch <- respData
}
