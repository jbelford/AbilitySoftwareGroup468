package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	//CTYPE   = "C_type"
<<<<<<< HEAD
	USER        = "UserId"
	AMOUNT      = "Amount"
	STOCK       = "StockSymbol"
	FILE        = "FileName"
	WEB_URL     = "http://webserver.prod.ability.com:44420"
	NUM_WORKERS = 10000
=======
	USER    = "UserId"
	AMOUNT  = "Amount"
	STOCK   = "StockSymbol"
	FILE    = "FileName"
	WEB_URL = "http://web:44420"
>>>>>>> 88af426bc82d2285c224b31fce52096be3f9e014
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

	c := make(chan endpoint, 100)
	done := make(chan bool, NUM_WORKERS)
	for i := 0; i < NUM_WORKERS; i++ {
		go startWorker(c, done)
	}

	log.Println("Sending Traffic to: " + WEB_URL)
	start := time.Now()

	// Request rate per second
	rate := 3000
	sleepTime := time.Second.Nanoseconds() / int64(rate)
	for _, req := range linesInFiles {
		c <- req
		time.Sleep(time.Duration(sleepTime))
	}

	close(c)
	for i := 0; i < NUM_WORKERS; i++ {
		<-done
	}

	now := time.Now()
	log.Println("Finished for loop")
	log.Println("Ran for: ", now.Sub(start))
	//log.Println(sentMessages)
}

func startWorker(ch chan endpoint, done chan bool) {
	for {
		data, ok := <-ch
		if !ok {
			done <- true
			return
		}
		log.Println("Sending", data)

		client := &http.Client{}
		req, err := http.NewRequest(data.Method, data.Query, nil)
		if err != nil {
			log.Println(err)
			return
		}
		req.Close = true
		req.Header.Set("Content-Type", "application/json")

		success := false
		for !success {
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			success = true
			if resp.StatusCode != 200 {
				log.Println(resp.StatusCode, data.Method, data.Query, resp.Body)
			}
			resp.Body.Close()
		}
	}
}

// rate := uint64(100)
// duration := 30 * time.Second
// targeter := vegeta.NewStaticTargeter(vegeta.Target{
// 	Method: "GET",
// 	URL:    "http://127.0.0.1:8081/test/quote?stock=ABC",
// })
// attacker := vegeta.NewAttacker()
// var metrics vegeta.Metrics
// for res := range attacker.Attack(targeter, rate, duration) {
// 	metrics.Add(res)
// }
// metrics.Close()
// fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
// reporter := vegeta.NewJSONReporter(&metrics)

// f, _ := os.OpenFile("results.json", os.O_WRONLY|os.O_CREATE, 0700)

// reporter.Report(f)

// f.Close()
