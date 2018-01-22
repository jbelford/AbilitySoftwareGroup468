package main

import (
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
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
	log.Print(com)
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Print(com)
	}
	writeResponse, err := conn.Write(com.commandObjToString())
	if err != nil {
		log.Print(com)
	}

	readResponse, err := conn.Read()
	if err != nil {
		log.Print(com)
	}
	conn.Close()

}

func userHandler(w http.ResponseWriter, r *http.Request) {

	user_info := common.CommandConstructor(r.FormValue("data"))
	passInfo(user_info)
	t := template.New("test.html")
	t, _ = t.ParseFiles("test.html")
	t.Execute(w, "")
}

func main() {
	log.Print("starts???")

	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", wrapHandler(userHandler)).Methods("GET")

	r.HandleFunc("/user", wrapHandler(userHandler)).Methods("POST", "GET")

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
