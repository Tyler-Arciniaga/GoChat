package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Hub struct {
	port string
	broadcastChannel chan Message
	joinChannel chan Client
	leaveChannel chan Client
	connectionMap map[net.Conn]Client
}

func (h Hub) Start(){
	fmt.Println("Listening on port", h.port)

	ln, err := net.Listen("tcp", h.port)
	if err != nil{
		fmt.Println("error creating network listener")
		os.Exit(1)
	}

	go h.HandleConnectionMap()

	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Println("error accepting new user connection", err)
			continue
		}

		go h.HandleConnection(conn)
	}
}

func (h Hub) HandleConnectionMap(){
	for {
		select{
		case m := <- h.broadcastChannel:
			h.BroadCastMessage(m)
		case j := <- h.joinChannel:
			h.connectionMap[j.Conn] = j
			go func(){
				h.broadcastChannel <- Message{Name: "Server", Msg: fmt.Sprintf("%s has joined the chat server\n", j.Name)}
			}()
		case l := <- h.leaveChannel:
			delete(h.connectionMap, l.Conn)
			go func(){
				h.broadcastChannel <- Message{Name: "Server", Msg: fmt.Sprintf("%s has disconnected from chat server\n", l.Name)}
			}()
		}
	}
}

func (h Hub) BroadCastMessage(m Message){
	for _, client := range h.connectionMap{
		go func(){
			client.MailBoxChan <- m
		}()
	}
}

func (h Hub) HandleConnection(conn net.Conn){
	client, err := h.CreateNewClient(conn)
	if err != nil {
		fmt.Println("error creating new client:", err)
	}

	h.joinChannel <- client
	
	go client.RecieveMessages(h.leaveChannel)
	go client.SendMessages(h.broadcastChannel, h.leaveChannel)
} //TODO: when to close connection?

func (h Hub) DisconnectClient(c Client){
	h.leaveChannel <- c
}

func (h Hub) CreateNewClient(conn net.Conn) (Client, error){
	tempReader := bufio.NewReader(conn)

	for {
		_, err := conn.Write([]byte("Enter name before entering chat room: "))
		if err != nil{
			return Client{}, errors.New("error prompting user for name")
		}

		bytes, err := tempReader.ReadBytes(byte('\n'))
		if err != nil{
			if err != io.EOF{
				return Client{}, errors.New("error reading in bytes when prompting for name")
			}
		}

		username := strings.TrimSpace(string(bytes))
		if len(username) < 1{
			continue
		}
		
		newMBC := make(chan Message)
		return Client{Conn: conn, Name: username, MailBoxChan: newMBC}, nil
	}
}