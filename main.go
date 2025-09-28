package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Chatter struct{
	Conn net.Conn
	Name string
} //should fields be exportable???

func main(){
	if len(os.Args) < 2{
		fmt.Println("Usage: go run main.go <port>")
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
	jChan := make(chan Chatter)
	lChan := make(chan net.Conn)

	connMap := make(map[net.Conn]string)
	go handleConnMap(connMap, bChan, jChan, lChan)

	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Println("error accepting new user connection", err)
			os.Exit(1)
		}

		newChatter := Chatter{Conn: conn, Name: "?"}
		jChan <- newChatter
		go handleConnection(conn, bChan, jChan, lChan)
		
	}
}

func handleConnection(conn net.Conn,  broadcastChannel chan string, joinChannel chan Chatter, leaveChannel chan net.Conn){
	defer conn.Close()
	bufReader := bufio.NewReader(conn)

	var username string

	for {
		_, err := conn.Write([]byte("Enter name before entering chat room: "))
		if err != nil{
			fmt.Println("error prompting user for name:", err)
		}
		bytes, err := bufReader.ReadBytes(byte('\n'))
		if err != nil{
			if err != io.EOF{
				fmt.Println("error reading in bytes from client", err)
			}
			return //triggers conn.Close()
		}
		username = strings.TrimSpace(string(bytes))
		if len(username) > 0{
			nameUpdated := Chatter{Conn: conn, Name: username}
			joinChannel <- nameUpdated
			break
		}
	}

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
			leaveChannel <- conn
			return //exit connection
		}
		res := fmt.Sprintf("[%s]: %s\n", username, strReq)

		//res := fmt.Sprintln("you said: ", strReq)
		fmt.Printf("server response: %s\n", res)

		
		broadcastChannel <- res
	}
}

func handleConnMap(connMap map[net.Conn]string, broadcastChannel chan string, joinChannel chan Chatter, leaveChannel chan net.Conn){
	for {
		select{
		case m := <- broadcastChannel:
			for conn, name := range connMap{
				if name == "?"{
					continue
				} //if user has not specified their name yet, do not present other user messages yet
				_, err := conn.Write([]byte(m))
				if err != nil{
					fmt.Println("error writing response back through conn:", err)
					return
				}
			}
		case chatter := <- joinChannel:
			connMap[chatter.Conn] = chatter.Name

		case conn := <- leaveChannel:
			delete(connMap, conn)
		}
	}
}