package main

import (
  "net"
  "encoding/xml"
  "log"
  "bufio"
  "os"
  //"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type AuditServer struct{}

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
