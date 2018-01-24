package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type WebServer struct{}

func (ws *WebServer) Start() {
	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/{user_id}/add", wrapHandler(userAddHandler)).Methods("POST")

	// r.HandleFunc("/user", wrapHandler(userHandler)).Methods("POST", "GET")

	// r.HandleFunc("/admin/DUMPLOG", wrapHandler(userHandler)).Methods("GET") //overall dumplog, try to not call this

	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(dir))))

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// user_info := common.CommandConstructor(r.FormValue("data"))
	// passInfo(user_info)
	t := template.New("test.html")
	t, _ = t.ParseFiles("./templates/test.html")
	t.Execute(w, "")
}

func userAddHandler(w http.ResponseWriter, r *http.Request) interface{} {
	amount, err := strconv.ParseFloat(r.URL.Query().Get("amount"), 32)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return common.Response{Success: false, Message: "Could not process field: 'amount'"}
	} else if amount <= 0 {
		return common.Response{Success: false, Message: "Parameter: 'amount' must be greater than 0"}
	}
	cmd := common.Command{
		C_type:    common.ADD,
		UserId:    mux.Vars(r)["user_id"],
		Amount:    int(amount * 100),
		Timestamp: time.Now(),
	}

	resp := issueTransactionCommand(cmd)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return common.Response{Success: false, Message: "Internal error prevented operation"}
	} else if !resp.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return resp
}

func issueTransactionCommand(com common.Command) *common.Response {
	textCmd, err := json.Marshal(com)
	if err != nil {
		return nil
	}

	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		return nil
	}
	defer conn.Close()

	_, err = conn.Write(append(textCmd, '\n'))

	var resp string
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	resp, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil
	}

	var jsonResp *common.Response
	err = json.Unmarshal([]byte(resp), &jsonResp)
	if err != nil {
		return nil
	}
	return jsonResp
}

func wrapHandler(
	handler func(w http.ResponseWriter, r *http.Request) interface{},
) func(w http.ResponseWriter, r *http.Request) {

	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// test input here/validity of requester
		resp := handler(w, r)

		respJSON, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write(respJSON)
		}
	}
	return h
}
