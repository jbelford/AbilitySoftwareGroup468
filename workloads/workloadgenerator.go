package main

import (
	"bufio"
	"log"
	"net/http"
	//"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"encoding/json"
	"bytes"
)

const (
	CTYPE = "C_type"
	USER = "UserId"
	AMOUNT = "Amount"
	STOCK = "StockSymbol"
	FILE = "FileName"
	WEB_URL = "http://webserver.prod.ability.com:44420"
)

func parseWorkloadCommand(cmdLine string) (string, string) {
	split_cmd := strings.Split(cmdLine, ",")
	if len(split_cmd) == 0 {
		log.Fatal("Empty Command!")
	}
	var WorkResp map[string]string

	C_type := split_cmd[0]

	end_url :=  WEB_URL + "/" + split_cmd[1] + "/" + strings.ToLower(C_type)

	switch C_type {
		case "ADD":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], AMOUNT: split_cmd[2]}
		case "QUOTE":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2]}
		case "BUY":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2], AMOUNT: split_cmd[3]}
		case "COMMIT_BUY":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1]}
		case "COMMIT_SELL":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1]}
		case "SELL":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2], AMOUNT: split_cmd[3]}
		case "CANCEL_SET_SELL":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2]}
		case "CANCEL_SET_BUY":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2]}
		case "SET_SELL_AMOUNT":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2], AMOUNT: split_cmd[3]}
		case "CANCEL_BUY":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1]}
		case "CANCEL_SELL":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1]}
		case "DISPLAY_SUMMARY":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1]}
		case "SET_BUY_AMOUNT":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2], AMOUNT: split_cmd[3]}
		case "SET_SELL_TRIGGER":
			WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], STOCK: split_cmd[2], AMOUNT: split_cmd[3]}
		case "DUMPLOG":
			if len(split_cmd) == 3 {
				// ADMIN DUMPLOG
				WorkResp = map[string]string{CTYPE: C_type, USER: split_cmd[1], FILE: split_cmd[2]}
				end_url = WEB_URL + "/" + "0" + "/" + strings.ToLower(C_type)
			} else {
				WorkResp = map[string]string{CTYPE: C_type, FILE: split_cmd[1]}
			}
		default:
			panic("unrecognized command: "+  C_type)
	}
	resp, _ := json.Marshal(WorkResp)
	return end_url, string(resp)
}

func main() {
	if len(os.Args) < 3 {
		panic("Missing arguments: <workLoadFile> <threadCount>")
	}
	pathToFile := os.Args[1]
	threadCount, _ := strconv.ParseInt(os.Args[2], 10, 0)
	file, err := os.Open(pathToFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var linesInFiles [][]string
	for scanner.Scan() {
		line := scanner.Text()
		endpoint, parsed := parseWorkloadCommand(strings.Split(line, "] ")[1])
		linesInFiles = append(linesInFiles, []string{endpoint,parsed})
	}

	sliceLength := len(linesInFiles)
	//log.Println(linesInFiles)
	log.Println("Sending:", sliceLength, "commands.")

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Start up threads to hit server...
	var wg sync.WaitGroup
	wg.Add(int(threadCount))
	sentMessages := make([]int, int(threadCount))


	log.Println("Sending Traffic to: " + WEB_URL + " using " + string(threadCount) + " threads...")
	start := time.Now()

	for i := 0; i < int(threadCount); i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < int(sliceLength); j++ {
				json_data := linesInFiles[j]
				b := new(bytes.Buffer)
				json.NewEncoder(b).Encode(json_data[1])
				_, err := http.Post(json_data[0], "application/json; charset=utf-8", b)

				if err != nil {
					// handle error
					log.Println(err)
				}

				sentMessages[i]++
			}
		}(i)
	}

	wg.Wait()
	now := time.Now()
	log.Println("Finished for loop")
	log.Println("Ran for: ", now.Sub(start))
	//log.Println(sentMessages)
}
