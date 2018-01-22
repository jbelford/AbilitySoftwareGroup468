package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/AbilitySoftwareGroup468/common"
	"github.com/gorilla/mux"
)

func wrapHandler(
	handler func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {

	h := func(w http.ResponseWriter, r *http.Request) {
		// test input here/validity of requester
		handler(w, r)
	}
	return h
}

//use pipes?
func passInfo(com common.Command) {
	conn, err := net.Dial("tcp", "192.168.1.134:80")
	if err != nil {
		log.Print(com)
		// handle error
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	log.Print(status)
	writeResponse, err := conn.Write(com.commandObjToString())
	log.Print(err)
	readResponse, err2 := conn.Read()
	log.Print(err2)

	conn.Close()
}

func userHandler(w http.ResponseWriter, r *http.Request) {

	temp_c_type, _ := strconv.Atoi(r.FormValue("c_type"))
	temp_amount, _ := strconv.Atoi(r.FormValue("amount"))
	user_info := common.Command{
		C_type:      temp_c_type,
		UserId:      r.FormValue("userid"),
		Amount:      temp_amount,
		StockSymbol: r.FormValue("stockSymbol"),
		FileName:    r.FormValue("fileName"),
	} //may not be necessary if it is handed as a parsed string
	//passInfo(user_info)
	t := template.New("test.html")
	t, _ = t.ParseFiles("test.html")
	t.Execute(w, vars)
}

func main() {
	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", wrapHandler(userHandler)).Methods("GET")

	r.HandleFunc("/user", wrapHandler(userHandler)).Methods("POST")

	r.HandleFunc("/admin/DUMPLOG", wrapHandler(userHandler)).Methods("GET") //overall dumplog, try to not call this

	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(dir))))

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}
