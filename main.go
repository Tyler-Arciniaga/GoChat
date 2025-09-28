package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Message struct{
	msg string
	sender string
}

func main(){
	if len(os.Args) < 3{
		fmt.Println("Usage: go run main.go <port> <username>")
		os.Exit(1)
	}
	port := fmt.Sprintf(":%s", os.Args[1])
	username := os.Args[2]

	ln, err := net.Listen("tcp", port)
	if err != nil{
		fmt.Println("error creating network listener")
		os.Exit(1)
	}

	fmt.Println("Listening on port", port)

	//bChan := make(chan Message)

	connMap := make(map[net.Conn]bool)
	go handleConnMap(connMap)

	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Println("error accepting new user connection", err)
			os.Exit(1)
		}

		go handleConnection(conn, username)
		
	}
}

func handleConnection(conn net.Conn, username string){
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
		res := fmt.Sprintf("[%s]: %s", username, strReq)

		//res := fmt.Sprintln("you said: ", strReq)
		//fmt.Printf("server response: %s\n", res)

		_, err = conn.Write([]byte(res))
		if err != nil{
			fmt.Println("error writing response back through conn:", err)
			return
		}
	}
}

func handleConnMap(connMap map[net.Conn]bool){
	for{
		select{

		}
	}
}