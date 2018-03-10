package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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

var MAX_WORKERS = uint64(8000)

type endpoint struct {
	Key    string
	Method string
	Query  string
}

type Result struct {
	Key       string
	Code      uint16        `json:"code"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	BytesOut  uint64        `json:"bytes_out"`
	BytesIn   uint64        `json:"bytes_in"`
	Error     string        `json:"error"`
}

var rest = map[string]endpoint{
	"ADD": endpoint{
		Key:    "ADD",
		Method: "POST",
		Query:  "%s/%d/%s/add?amount=%s",
	},
	"QUOTE": endpoint{
		Key:    "QUOTE",
		Method: "GET",
		Query:  "%s/%d/%s/quote?stock=%s",
	},
	"BUY": endpoint{
		Key:    "BUY",
		Method: "POST",
		Query:  "%s/%d/%s/buy?stock=%s&amount=%s",
	},
	"COMMIT_BUY": endpoint{
		Key:    "COMMIT_BUY",
		Method: "POST",
		Query:  "%s/%d/%s/commit_buy",
	},
	"CANCEL_BUY": endpoint{
		Key:    "CANCEL_BUY",
		Method: "POST",
		Query:  "%s/%d/%s/cancel_buy",
	},
	"SELL": endpoint{
		Key:    "SELL",
		Method: "POST",
		Query:  "%s/%d/%s/sell?stock=%s&amount=%s",
	},
	"COMMIT_SELL": endpoint{
		Key:    "COMMIT_SELL",
		Method: "POST",
		Query:  "%s/%d/%s/commit_sell",
	},
	"CANCEL_SELL": endpoint{
		Key:    "CANCEL_SELL",
		Method: "POST",
		Query:  "%s/%d/%s/cancel_sell",
	},
	"SET_BUY_AMOUNT": endpoint{
		Key:    "SET_BUY_AMOUNT",
		Method: "POST",
		Query:  "%s/%d/%s/set_buy_amount?stock=%s&amount=%s",
	},
	"CANCEL_SET_BUY": endpoint{
		Key:    "CANCEL_SET_BUY",
		Method: "POST",
		Query:  "%s/%d/%s/cancel_set_buy?stock=%s",
	},
	"SET_BUY_TRIGGER": endpoint{
		Key:    "SET_BUY_TRIGGER",
		Method: "POST",
		Query:  "%s/%d/%s/set_buy_trigger?stock=%s&amount=%s",
	},
	"SET_SELL_AMOUNT": endpoint{
		Key:    "SET_SELL_AMOUNT",
		Method: "POST",
		Query:  "%s/%d/%s/set_sell_amount?stock=%s&amount=%s",
	},
	"SET_SELL_TRIGGER": endpoint{
		Key:    "SET_SELL_TRIGGER",
		Method: "POST",
		Query:  "%s/%d/%s/set_sell_trigger?stock=%s&amount=%s",
	},
	"CANCEL_SET_SELL": endpoint{
		Key:    "CANCEL_SET_SELL",
		Method: "POST",
		Query:  "%s/%d/%s/cancel_set_sell?stock=%s",
	},
	"DUMPLOG": endpoint{
		Key:    "DUMPLOG",
		Method: "GET",
		Query:  "%s/%d/%s/dumplog?filename=%s",
	},
	"DISPLAY_SUMMARY": endpoint{
		Key:    "DISPLAY_SUMMARY",
		Method: "GET",
		Query:  "%s/%d/%s/display_summary",
	},
}

func parseWorkloadCommand(cmdLine string, i int64) endpoint {
	split_cmd := strings.Split(cmdLine, ",")
	if len(split_cmd) == 0 {
		log.Fatal("Empty Command!")
	}

	mapped := rest[split_cmd[0]]
	split := make([]interface{}, len(split_cmd)+1)
	split[0] = WEB_URL
	split[1] = i
	for i, val := range split_cmd[1:] {
		temp_val := strings.TrimSpace(val)
		amount, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err == nil { // It's a float!
			temp_val = strconv.Itoa(int(amount * 100.0))
		}
		split[i+2] = temp_val
	}
	if split_cmd[0] == "DUMPLOG" && len(split) < 4 {
		mapped.Query = fmt.Sprintf(mapped.Query, split[0], i, "admin", split[2])
	} else {
		mapped.Query = fmt.Sprintf(mapped.Query, split...)
	}

	return mapped
}

func main() {
	if len(os.Args) < 2 {
		panic("Missing arguments: <workLoadFile>")
	} else if len(os.Args) == 3 {
		MAX_WORKERS, _ = strconv.ParseUint(os.Args[2], 10, 64)
	}
	log.Printf("Using %d workers...", MAX_WORKERS)

	pathToFile := os.Args[1]
	file, err := os.Open(pathToFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	i := 1
	scanner := bufio.NewScanner(file)

	var linesInFiles []endpoint
	for scanner.Scan() {
		line := scanner.Text()
		endpoint := parseWorkloadCommand(strings.Split(line, "] ")[1], int64(i))
		i++
		linesInFiles = append(linesInFiles, endpoint)
	}

	gen := NewGenerator(1000)

	log.Println("Sending Traffic to: " + WEB_URL)
	start := time.Now()

	rate := uint64(3000)

	results := []*Result{}
	for r := range gen.Start(linesInFiles, rate) {
		results = append(results, r)
	}

	now := time.Now()
	log.Println("Finished for loop")
	log.Println("Ran for: ", now.Sub(start))
	log.Println(fmt.Sprintf("Sent %d requests at a rate of %d req/s", len(linesInFiles), rate))

	printStats(results)
}

type Statistics struct {
	Count          uint64
	AverageLatency time.Duration
	TotalBytesOut  uint64
	TotalBytesIn   uint64
	Errors         []string
}

func printStats(results []*Result) {
	msg := "--------------------\nSTATISTICs\n--------------------"
	stats := make(map[string]Statistics)
	for _, r := range results {
		s := stats[r.Key]
		s.Count++
		s.AverageLatency += r.Latency
		s.TotalBytesIn += r.BytesIn
		s.TotalBytesOut += r.BytesOut
		if r.Error != "" {
			s.Errors = append(s.Errors, r.Error)
		}
		stats[r.Key] = s
	}
	for k, s := range stats {
		// Fix the latency calculation
		latency := float64(int64(s.AverageLatency/time.Millisecond)) / float64(s.Count)
		s.AverageLatency = s.AverageLatency / time.Duration(s.Count)
		msg += fmt.Sprintf("\n\n%s - Average Latency: %f ms - Total Bytes out: %d - Total Bytes in: %d - Errors: %d\n",
			k, latency, s.TotalBytesOut, s.TotalBytesIn, len(s.Errors))
		for _, e := range s.Errors {
			msg += "\t" + e + "\n"
		}
	}
	if f, err := os.OpenFile("stats.txt", os.O_WRONLY|os.O_CREATE, 0777); err == nil {
		defer f.Close()
		f.WriteString(msg)
	} else {
		log.Println(msg)
	}
}

type ReqInfo struct {
	Timestamp time.Time
	Endpoint  endpoint
}

type WorkLoadGenerator struct {
	workers uint64
	client  http.Client
	dialer  *net.Dialer
}

func NewGenerator(workers uint64) *WorkLoadGenerator {
	wl := &WorkLoadGenerator{workers: workers}
	localAddr := net.IPAddr{IP: net.IPv4zero}
	timeout := 30 * time.Second
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	wl.dialer = &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: localAddr.IP, Zone: localAddr.Zone},
		KeepAlive: 30 * time.Second,
		Timeout:   timeout,
	}
	wl.client = http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial:  wl.dialer.Dial,
			ResponseHeaderTimeout: timeout,
			TLSClientConfig:       tlsConfig,
			TLSHandshakeTimeout:   10 * time.Second,
			MaxIdleConnsPerHost:   10000,
		},
	}
	return wl
}

func (wl *WorkLoadGenerator) Start(work []endpoint, rate uint64) chan *Result {
	var workers sync.WaitGroup
	results := make(chan *Result)

	ticks := make(chan ReqInfo)
	for i := uint64(0); i < wl.workers; i++ {
		workers.Add(1)
		go wl.attack(&workers, ticks, results)
	}

	go func() {
		defer close(results)
		defer workers.Wait()
		defer close(ticks)
		interval := 1e9 / rate
		hits := uint64(len(work))
		began, done := time.Now(), uint64(0)
		for {
			now, next := time.Now(), began.Add(time.Duration(done*interval))
			time.Sleep(next.Sub(now))
			select {
			case ticks <- ReqInfo{Timestamp: max(next, now), Endpoint: work[done]}:
				if done++; done == hits {
					return
				}
			default: // All workers are blocked so lets create another worker
				if wl.workers < MAX_WORKERS {
					wl.workers++
					workers.Add(1)
					go wl.attack(&workers, ticks, results)
				}
			}
		}
	}()

	return results
}

func (wl *WorkLoadGenerator) attack(workers *sync.WaitGroup, ticks chan ReqInfo, results chan *Result) {
	defer workers.Done()
	for reqInfo := range ticks {
		results <- wl.hit(reqInfo)
	}
}

func (wl *WorkLoadGenerator) hit(reqInfo ReqInfo) *Result {
	var err error
	ep := reqInfo.Endpoint
	res := Result{Timestamp: reqInfo.Timestamp, Key: ep.Key}

	defer func() {
		res.Latency = time.Since(reqInfo.Timestamp)
		if err != nil {
			res.Error = err.Error()
		}
	}()

	req, err := http.NewRequest(ep.Method, ep.Query, nil)
	if err != nil {
		return &res
	}

	r, err := wl.client.Do(req)
	if err != nil {
		return &res
	}
	defer r.Body.Close()

	in, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		return &res
	}
	res.BytesIn = uint64(in)

	if req.ContentLength != -1 {
		res.BytesOut = uint64(req.ContentLength)
	}

	if res.Code = uint16(r.StatusCode); res.Code < 200 || res.Code >= 400 {
		res.Error = r.Status
	}

	return &res
}

func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
