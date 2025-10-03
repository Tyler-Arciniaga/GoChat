package main

import (
	"fmt"
	"os"
)


func (m Message) String() string{
	return fmt.Sprintf("[%s]: %s\n", m.Name, m.Msg)
}

func main(){
	if len(os.Args) < 2{
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := fmt.Sprintf(":%s", os.Args[1])
	joinChan := make(chan Client)
	leaveChan := make(chan Client)
	roomMap := make(map[int]*Room)
	clientMap := make(map[string]Client)

	
	hub := Hub{port: port, joinChannel: joinChan, leaveChannel: leaveChan, roomMap: roomMap, clientMap: clientMap}
	hub.Start()
}