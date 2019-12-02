package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

const (
	cREPLY = iota
	cREGISTER
	cNOTIFY
	cCHECK
)

type tMsg struct {
	Code  int
	Addr  string
	Addrs []string
	Num   int
}

func main() {
	var remoteAddr string
	msg := tMsg{cCHECK, "not important", []string{}, 0}
	for {
		fmt.Print("Remote: ")
		fmt.Scanln(&remoteAddr)
		if remoteAddr == "." { break }
		for {
			fmt.Print("Num: ")
			fmt.Scanf("%d\n", &msg.Num)
			if msg.Num < 0 { break }
			sendRec(remoteAddr, msg, func(c net.Conn) {
				dec := json.NewDecoder(c)
				var msg tMsg
				dec.Decode(&msg)
				if msg.Num < 0 {
					fmt.Println("Unknown")
				} else {
					fmt.Println("Known!")
				}
			})
		}
	}
}
func sendRec(remoteAddr string, msg tMsg, resp func(c net.Conn)) {
	if conn, err := net.Dial("tcp", remoteAddr); err != nil {
		log.Println(err.Error())
	} else {
		defer conn.Close()
		enc := json.NewEncoder(conn)
		fmt.Println("Sending", msg, "to", remoteAddr)
		enc.Encode(&msg)
		if resp != nil {
			resp(conn)
		}
	}
}