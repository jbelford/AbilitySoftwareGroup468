package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//CTYPE   = "C_type"
	USER    = "UserId"
	AMOUNT  = "Amount"
	STOCK   = "StockSymbol"
	FILE    = "FileName"
	WEB_URL = "http://web:44420"
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
		Query:  "%s/%s/quote?stock=%s",
	},
	"BUY": endpoint{
		Method: "POST",
		Query:  "%s/%s/buy?stock=%s&amount=%s",
	},
	"COMMIT_BUY": endpoint{
		Method: "POST",
		Query:  "%s/%s/commit_buy",
	},
	"CANCEL_BUY": endpoint{
		Method: "POST",
		Query:  "%s/%s/cancel_buy",
	},
	"SELL": endpoint{
		Method: "POST",
		Query:  "%s/%s/sell?stock=%s&amount=%s",
	},
	"COMMIT_SELL": endpoint{
		Method: "POST",
		Query:  "%s/%s/commit_sell",
	},
	"CANCEL_SELL": endpoint{
		Method: "POST",
		Query:  "%s/%s/cancel_sell",
	},
	"SET_BUY_AMOUNT": endpoint{
		Method: "POST",
		Query:  "%s/%s/set_buy_amount?stock=%s&amount=%s",
	},
	"CANCEL_SET_BUY": endpoint{
		Method: "POST",
		Query:  "%s/%s/cancel_set_buy?stock=%s",
	},
	"SET_BUY_TRIGGER": endpoint{
		Method: "POST",
		Query:  "%s/%s/set_buy_trigger?stock=%s&amount=%s",
	},
	"SET_SELL_AMOUNT": endpoint{
		Method: "POST",
		Query:  "%s/%s/set_sell_amount?stock=%s&amount=%s",
	},
	"SET_SELL_TRIGGER": endpoint{
		Method: "POST",
		Query:  "%s/%s/set_sell_trigger?stock=%s&amount=%s",
	},
	"CANCEL_SET_SELL": endpoint{
		Method: "POST",
		Query:  "%s/%s/cancel_set_sell?stock=%s",
	},
	"DUMPLOG": endpoint{
		Method: "GET",
		Query:  "%s/%s/dumplog?filename=%s",
	},
	"DISPLAY_SUMMARY": endpoint{
		Method: "GET",
		Query:  "%s/%s/display_summary",
	},
}

func parseWorkloadCommand(cmdLine string) endpoint {
	split_cmd := strings.Split(cmdLine, ",")
	if len(split_cmd) == 0 {
		log.Fatal("Empty Command!")
	}
	//var WorkResp map[string]string

	//C_type := split_cmd[0]

	mapped := rest[split_cmd[0]]
	split := make([]interface{}, len(split_cmd))
	split[0] = WEB_URL
	for i, val := range split_cmd[1:] {
		temp_val := strings.TrimSpace(val)
		amount, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err == nil { // It's a float!
			temp_val = strconv.Itoa(int(amount * 100.0))
		}
		split[i+1] = temp_val
	}
	if split_cmd[0] == "DUMPLOG" && len(split) < 3 {
		mapped.Query = fmt.Sprintf(mapped.Query, split[0], "admin", split[1])
	} else {
		mapped.Query = fmt.Sprintf(mapped.Query, split...)
	}

	return mapped
}

func main() {
	if len(os.Args) < 2 {
		panic("Missing arguments: <workLoadFile>")
	}
	pathToFile := os.Args[1]
	file, err := os.Open(pathToFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var linesInFiles []endpoint
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
	// wg.Add(int(threadCount))
	// sentMessages := make([]int, int(threadCount))

	log.Println("Sending Traffic to: " + WEB_URL)
	start := time.Now()

	for i := 0; i < sliceLength; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			json_data := linesInFiles[j]
			log.Println(j, "Sending", json_data, json_data)

			if json_data.Method == "POST" {
				resp, err := http.PostForm(json_data.Query, url.Values{})
				if err != nil {
					// handle error
					log.Println(err)
					return
				}

				if resp.StatusCode != 200 {
					log.Println(json_data, resp)
				}
			} else {
				resp, err := http.Get(json_data.Query)
				if err != nil {
					// handle error
					log.Println(err)
					return
				}

				if resp.StatusCode != 200 {
					log.Println(json_data, resp)
				}
			}

			// sentMessages[i]++
		}(i)
		if (i+1)%100 == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
	now := time.Now()
	log.Println("Finished for loop")
	log.Println("Ran for: ", now.Sub(start))
	//log.Println(sentMessages)
}
