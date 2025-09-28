package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main(){
	if len(os.Args) < 2{
		fmt.Println("Usage: go run main.go <port> <username>")
		os.Exit(1)
	}
	port := fmt.Sprintf(":%s", os.Args[1])

	ln, err := net.Listen("tcp", port)
	if err != nil{
		fmt.Println("error creating network listener")
		os.Exit(1)
	}

	fmt.Println("Listening on port", port)

	bChan := make(chan string)
	uChan := make(chan net.Conn)

	connMap := make(map[net.Conn]bool)
	go handleConnMap(connMap, bChan, uChan)

	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Println("error accepting new user connection", err)
			os.Exit(1)
		}

		
		uChan <- conn
		go handleConnection(conn, bChan)
		
	}
}

func handleConnection(conn net.Conn,  broadcastChannel chan string){
	defer conn.Close()

	bufReader := bufio.NewReader(conn)
	for {
		bytes, err := bufReader.ReadBytes(byte('\n'))
		if err != nil{
			if err != io.EOF{
				fmt.Println("error reading in bytes from client", err)
			}
			return //triggers conn.Close()
		}

		strReq := strings.TrimSpace(string(bytes))
		if strReq == "/leave"{
			return //exit connection
		}
		res := fmt.Sprintf("[some_user]: %s\n", strReq)

		//res := fmt.Sprintln("you said: ", strReq)
		fmt.Printf("server response: %s\n", res)

		
		broadcastChannel <- res
	}
}

func handleConnMap(connMap map[net.Conn]bool, broadcastChannel chan string, userChannel chan net.Conn){
	for {
		select{
		case m := <- broadcastChannel:
			for conn := range connMap{
				_, err := conn.Write([]byte(m))
				if err != nil{
					fmt.Println("error writing response back through conn:", err)
					return
				}
			}
		case c := <- userChannel:
			connMap[c] = true
		}
	}
}