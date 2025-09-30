package main

import (
	"fmt"
	"os"
)

type MessageType int

const (
	Broadcast MessageType = iota
	Whisper
)

type Message struct{
	Type MessageType
	Name string
	To string
	Msg string
}

func (m Message) String() string{
	return fmt.Sprintf("[%s]: %s\n", m.Name, m.Msg)
}

func main(){
	if len(os.Args) < 2{
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := fmt.Sprintf(":%s", os.Args[1])
	broadcastChan := make(chan Message)
	joinChan := make(chan Client)
	leaveChan := make(chan Client)
	connMap := make(map[string]Client)

	
	hub := Hub{port: port, broadcastChannel: broadcastChan, joinChannel: joinChan, leaveChannel: leaveChan, connectionMap: connMap}
	hub.Start()
}