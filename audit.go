package main

import (
  "net"
  "encoding/xml"
  "log"
  "bufio"
  "os"
  "github.com/gorilla/mux"
  "github.com/gorilla/rpc/v2"
  "github.com/gorilla/rpc/v2/json"
  "net/http"
  "github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type AuditServer struct{}

type Logging struct {}

type Result string

func (l *Logging) logUserCommand(r *http.Request, args *common.Args, result *Result) error {
  return nil
}

func (l *Logging) logQuoteServer(r *http.Request, args *common.Args, result *Result) error {

    return nil
}

func (l *Logging) logAccountTransaction(r *http.Request, args *common.Args, result *Result) error {

    return nil
}

func (l *Logging) logSystemEvent(r *http.Request, args *common.Args, result *Result) error {

    return nil
}

func (l *Logging) logErrorEvent(r *http.Request, args *common.Args, result *Result) error {

    return nil
}

func (l *Logging) logDebugEvent(r *http.Request, args *common.Args, result *Result) error {

    return nil
}

func log_msg(MSG string) {
  log.Println("TODO:// Add user ID to log_msg struct")
  user := "1"

  user_log, err := xml.MarshalIndent(MSG, "  ", "    ")
  if err != nil {
    log.Println("error: %v\n", err)
  }

  f1, err := os.OpenFile(user + ".txt", os.O_APPEND|os.O_WRONLY, 0600)
  if err != nil {
    panic(err)
  }

  f2, err := os.OpenFile("all_users.txt", os.O_APPEND|os.O_WRONLY, 0600)
  if err != nil {
    panic(err)
  }

  defer f1.Close()
  defer f2.Close()

  enc := xml.NewEncoder(f1)
	enc.Indent("  ", "    ")
	enc.Encode(user_log)
  enc2 := xml.NewEncoder(f2)
  enc2.Encode(user_log)
}

func (ad *AuditServer) Start() {
  ln, err := net.Listen("tcp", "127.0.0.2:8081")
  if err != nil {
    log.Fatal(err)
  }

  server := rpc.NewServer()
  server.RegisterCodec(json.NewCodec(), "application/json")
  logging := new(Logging)
  server.RegisterService(logging, "")

  router := mux.NewRouter()
  router.Handle("/Log",server)
  log.Println(http.ListenAndServe(":44424",router))

  for {
    conn, err := ln.Accept()
    message, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
      continue
    }
    log.Println("Received: ", string(message))
    //defer log_msg("1", message)
    conn.Close()
  }
}
