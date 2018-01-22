package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
		parsed := strings.Split(line, "] ")[1]
		linesInFiles = append(linesInFiles, parsed)
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

	web_url := "http://127.0.0.1:8000"

	log.Println("Sending Traffic using " + string(threadCount) + " threads...")
	start := time.Now()

	for i := 0; i < int(threadCount); i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < int(sliceLength); j++ {
				line := linesInFiles[j]

				sent_data := url.Values{}
				sent_data.Set("data", line)

				_, err := http.PostForm(web_url+"/user", sent_data)

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
