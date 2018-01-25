package main

import (
  //"net"
  //"encoding/xml"

  //"github.com/mattpaletta/AbilitySoftwareGroup468/common"
)

type AuditServer struct{}

/*
v := &Person{Id: 13, FirstName: "John", LastName: "Doe", Age: 42}
output, err := xml.MarshalIndent(v, "  ", "    ")
  	if err != nil {
  		fmt.Printf("error: %v\n", err)
  	}

  	os.Stdout.Write(output)
*/

func (user string, MSG struct) log_msg() {
  user_log, err := xml.MarshalIndent(MSG, "  ", "    ")
  if err != nil {
    fmt.Printf("error: %v\n", err)
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

  if _, err = f1.WriteString(user_log); err != nil {
    panic(err)
  }

  if _, err = f2.WriteString(user_log); err != nil {
    panic(err)
  }
}

func (ad *AuditServer) Start() {
  ln, err := net.Listen("tcp", "127.0.0.1:8082")
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
    defer log_msg(message)
    conn.Close()
  }
}
