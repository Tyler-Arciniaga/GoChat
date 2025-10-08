package main

import (
	common "go-chat/Common"
	"net"
)

type Hub struct {
	port          string
	numRooms      int
	joinChannel   chan net.Conn
	leaveChannel  chan net.Conn
	clientRoomMap map[net.Conn]*Room
	roomMap       map[int]*Room
}

type Room struct {
	roomID           int
	chatterMap       map[string]net.Conn
	messageChannel   chan common.Message //recieves raw unparsed messages from hub, routes to appropriate internal channel
	joinChannel      chan AdminSignal
	leaveChannel     chan AdminSignal
	broadcastChannel chan common.Message
	whisperChannel   chan common.Message
}

type AdminSignal struct {
	message common.Message
	conn    net.Conn
}
