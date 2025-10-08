package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := fmt.Sprintf(":%s", os.Args[1])
	joinChan := make(chan net.Conn)
	leaveChan := make(chan net.Conn)
	clientRoomMap := make(map[net.Conn]*Room)

	hub := Hub{port: port, numRooms: 0, joinChannel: joinChan, leaveChannel: leaveChan, clientRoomMap: clientRoomMap}
	hub.Start()
}
