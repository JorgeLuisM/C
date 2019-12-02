package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
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

var localAddr string
var numbers map[int]string
var chAddrs chan []string

func main() {
	fmt.Print("Dirección local [ejemplo:puerto]: ")
	fmt.Scanln(&localAddr)
	genNumbers()
	connect2Next()
	server()
}
func genNumbers() {
	rand.Seed(time.Now().UTC().UnixNano())
	n := rand.Intn(10)
	numbers = make(map[int]string)
	for i := 0; i < n; i++ {
		numbers[rand.Intn(100)] = ""
	}
	fmt.Println(numbers)
}
func connect2Next() {
	chAddrs = make(chan []string)
	var remote string
	fmt.Print("Dirección remota [remota:puerto]: ")
	fmt.Scanln(&remote)
	if remote != "" {
		go sendRec(remote, tMsg{cREGISTER, localAddr, []string{}, 0},
			func(conn net.Conn) {
				var msg tMsg
				dec := json.NewDecoder(conn)
				dec.Decode(&msg)
				fmt.Println("Resp", msg)
				chAddrs <- append(msg.Addrs, remote)
			})
	} else {
		go func() { chAddrs <- make([]string, 0, 10) }()
	}
}
func server() {
	if ln, err := net.Listen("tcp", localAddr); err != nil {
		log.Panicln(err.Error())
	} else {
		defer ln.Close()
		fmt.Println(localAddr, "listening")
		for {
			if conn, err := ln.Accept(); err != nil {
				log.Panicln(err.Error())
			} else {
				go handle(conn)
			}
		}
	}
}
func handle(conn net.Conn) {
	defer conn.Close()
	fmt.Println(conn.RemoteAddr(), "accepted")
	var msg tMsg
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&msg); err != nil {
		log.Println(err.Error())
	} else {
		fmt.Println("Got", msg)
		switch msg.Code {
		case cREGISTER:
			register(conn, msg)
		case cNOTIFY:
			notify(msg)
		case cCHECK:
			check(conn, msg)
		}
	}
}
func register(conn net.Conn, msg tMsg) {
	addrs := <-chAddrs
	enc := json.NewEncoder(conn)
	enc.Encode(&tMsg{cREPLY, localAddr, addrs, 0})
	for _, addr := range addrs {
		send(addr, tMsg{cNOTIFY, msg.Addr, []string{}, 0})
	}
	go addAddr(addrs, msg.Addr)
}
func notify(msg tMsg) {
	addrs := <-chAddrs
	go addAddr(addrs, msg.Addr)
}
func check(conn net.Conn, msg tMsg) {
	enc := json.NewEncoder(conn)
	num := -1
	if _, ok := numbers[msg.Num]; ok {
		num = msg.Num
	}
	enc.Encode(&tMsg{cREPLY, localAddr, []string{}, num})
}
func addAddr(addrs []string, addr string) {
	chAddrs <- append(addrs, addr)
	fmt.Println(addr, "added")
}
func send(remoteAddr string, msg tMsg) {
	sendRec(remoteAddr, msg, nil)
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
