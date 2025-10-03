package main

import "net"

type MessageType int
const (
	Broadcast MessageType = iota
	Whisper
	Leave
)

type DisconnectionType int 
const (
	Interruption DisconnectionType = iota
	Graceful
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
	clientMap map[string]Client
}

type Client struct{
	Conn net.Conn
	Name string
	MailBoxChan chan Message
}

type LeaveSignal struct{
	LeaveType DisconnectionType
	Client Client
}

type Room struct{
	roomID int
	connectionMap map[string]Client
	messageChannel chan Message
	joinChannel chan *Client
	leaveChannel chan LeaveSignal
}