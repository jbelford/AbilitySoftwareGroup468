package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	//"net/url"
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CTYPE   = "C_type"
	USER    = "UserId"
	AMOUNT  = "Amount"
	STOCK   = "StockSymbol"
	FILE    = "FileName"
	WEB_URL = "http://webserver.prod.ability.com:44420"
)

type endpoint struct {
	Method string
	Query  string
}

var rest = map[string]endpoint{
	"ADD": endpoint{
		Method: "POST",
		Query:  "%s/%s/add?amount=%s",
	},
	"QUOTE": endpoint{
		Method: "GET",
		Query: "%s/%s/quote?stock=%s",
	},
	"BUY": endpoint{
		Method: "POST",
		Query: "%s/%s/buy?stock=%s&amount=%s",
	},
	"COMMIT_BUY": endpoint{
		Method: "POST",
		Query: "%s/%s/commit_buy",
	},
	"CANCEL_BUY": endpoint{
		Method: "POST",
		Query: "%s/%s/cancel_buy",
	},
	"SELL": endpoint{
		Method: "POST",
		Query: "%s/%s/sell?stock=%s&amount=%s"
	},
	"COMMIT_SELL": endpoint{
		Method: "POST",
		Query: "%s/%s/commit_sell",
	},
	"CANCEL_SELL": endpoint{
		Method: "POST",
		Query: "%s/%s/cancel_sell",
	},
	"SET_BUY_AMOUNT": endpoint{
		Method: "POST",
		Query: "%s/%s/set_buy_amount?stock=%s&amount=%s"
	},
	"CANCEL_SET_BUY": endpoint{
		Method: "POST",
		Query: "%s/%s/cancel_set_buy?stock=%s"
	},
	"SET_BUY_TRIGGER": endpoint{
		Method: "POST",
		Query: "%s/%s/set_buy_trigger?stock=%s&amount=%s",
	},
	"SET_SELL_AMOUNT": endpoint{
		Method: "POST",
		Query: "%s/%s/set_sell_amount?stock=%s&amount=%s"
	},
	"SET_SELL_TRIGGER": endpoint{
		Method: "POST",
		Query: "%s/%s/set_sell_trigger?stock=%s&amount=%s"
	},
	"CANCEL_SET_SELL": endpoint{
		Method: "POST",
		Query: "%s/%s/cancel_set_sell?stock=%s"
	},
	"DUMPLOG": endpoint{
		Method: "POST",
		Query: "%s/%s/dumplog?filename=%s"
	},
	"DUMPLOG": endpoint{
		Method: "POST",
		Query: "%s/0/dumplog?filename=%s"
	},
	"DISPLAY_SUMMARY": endpoint{
		Method: "GET",
		Query: "%s/%s/display_summary"
	}

}

func parseWorkloadCommand(cmdLine string) (string, string) {
	split_cmd := strings.Split(cmdLine, ",")
	if len(split_cmd) == 0 {
		log.Fatal("Empty Command!")
	}
	var WorkResp map[string]string

	C_type := split_cmd[0]

	mapped := rest[split_cmd[0]]
	split := make([]interface{}, 0)
	split = append(split, split_cmd[1:])
	end_url := fmt.Sprintf(mapped.Query, split...)
	//log.Println(end_url)

	return end_url
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

	var linesInFiles []string
	for scanner.Scan() {
		line := scanner.Text()
		endpoint := parseWorkloadCommand(strings.Split(line, "] ")[1])
		linesInFiles = append(linesInFiles, endpoint)
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
				log.Println("Sending", json_data, json_data)
				_, err := http.Post(json_data)

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
