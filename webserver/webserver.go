package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/AbilitySoftwareGroup468/common"

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

func passInfo(com common.Command){
	conn, err := net.Dial("tcp", "192.168.1.134:80")
	if err != nil {
		log.Print(com)
		// handle error
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
  status, err := bufio.NewReader(conn).ReadString('\n')
	common.Command
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user_info := common.Command({
		userid: vars["userid"]
		
	})
	passInfo(vars)
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

	r.HandleFunc("/user/{command}", wrapHandler(userHandler)).Methods("GET")

	r.HandleFunc("/admin/DUMPLOG", wrapHandler(userHandler)).Methods("GET") //overall dumplog, try to not call this

	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(dir))))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}
