package main

import "net"

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

type Hub struct {
	port string
	joinChannel chan Client
	leaveChannel chan Client
	roomMap map[int]*Room
	clientMap map[string]bool
}

type Client struct{
	Conn net.Conn
	Name string
	MailBoxChan chan Message
	ActiveRoomChan chan Message
	ActiveLeaveChan chan Client
}

type Room struct{
	roomID int
	connectionMap map[string]Client
	messageChannel chan Message
	joinChannel chan *Client
	leaveChannel chan Client
}