package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func (h Hub) Start(){
	defer func(){
		if v := recover(); v != nil{
			slog.Error("Unrecoverable panic detected, shutting down chat server now.", "panic", v)
			h.CleanUpHub()
		}
	}()
	fmt.Println("Listening on port", h.port)

	ln, err := net.Listen("tcp", h.port)
	if err != nil{
		fmt.Println("error creating network listener")
		os.Exit(1)
	}

	for i := range(3){
		newRoom := Room{roomID: i, connectionMap: make(map[string]Client), 
			messageChannel: make(chan Message), joinChannel: make(chan *Client), leaveChannel: make(chan LeaveSignal)}
		h.roomMap[i] = &newRoom
		go newRoom.HandleConnectionMap(h.leaveChannel)
	}

	go h.HandleClientMap()

	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Println("error accepting new user connection", err)
			continue
		}

		go h.HandleConnection(conn)
	}
}

func (h Hub) HandleClientMap(){
	for{
		select{
		case j := <- h.joinChannel:
			h.clientMap[strings.ToLower(j.Name)] = j
		case l := <- h.leaveChannel:
			delete(h.clientMap, strings.ToLower(l.Name))
			slog.Info(fmt.Sprintf("Client: %s has disconnected from Go Chat", l.Name))
		}
	}
}

func (h Hub) PromptRoomSelect(c Client) (*Room, error){
	tempReader := bufio.NewReader(c.Conn)
	for {
		_, err := c.Conn.Write([]byte("Pick a chat room (-1 to disconnect from sever):\n0\n1\n2\n-> ")) //TODO: refactor so that room numbers are not hardcoded
		if err != nil {
			return nil, fmt.Errorf("error writing to connection: %s", err)
		}

		bytes, err := tempReader.ReadBytes(byte('\n'))
		if err != nil {
			return nil, errors.New("error reading in bytes when prompting for room id")
		}
		
		roomChoiceStr := strings.TrimSpace(string(bytes))
		if roomChoiceStr == "-1"{
			c.Conn.Write([]byte("Come back soon :)\n"))
			h.leaveChannel <- c
			c.Conn.Close()
			return nil, nil
		}
		if len(roomChoiceStr) < 1{
			continue
		}
		roomChoiceInt, err := strconv.Atoi(roomChoiceStr)
		if err != nil {
			continue
		}

		room, ok := h.roomMap[roomChoiceInt]
		if !ok{
			_, err := c.Conn.Write([]byte("Invalid room number, try again\n"))
			if err != nil {
				return nil, fmt.Errorf("error writing to connection: %s", err)
			}
			continue
		}
		return room, nil
	}
}

func (h Hub) HandleConnection(conn net.Conn){
	client, err := h.CreateNewClient(conn)
	if err != nil {
		fmt.Println("error creating new client:", err)
		_, err := conn.Write([]byte("error connecting to server try again"))
		if err != nil {
			slog.Error(fmt.Sprint("error writing error message to connection", err))
		}
		conn.Close()
		return
	}
	
	h.joinChannel <- client

	for {
		room, err := h.PromptRoomSelect(client)
		if room == nil{
			return
		}// user has disconnected from chat server with /leave command
		if err != nil{
			slog.Error("error with room selection, booting client")
			h.leaveChannel <- client
			return
		}

		room.joinChannel <- &client

		var wg sync.WaitGroup
		wg.Add(2)

		go func(){
			defer wg.Done()
			client.SendMessages(room.messageChannel, room.leaveChannel)
		}()

		go func(){
			defer wg.Done()
			client.RecieveMessages(room.leaveChannel)
		}()

		wg.Wait()
	}
	
}

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

		_, exists := h.clientMap[strings.ToLower(username)]
		if exists{
			_, err = conn.Write([]byte("Name already taken in current chat room\n"))
			if err != nil {
				return Client{}, errors.New("error requesting new name")
			}
			continue
		}
		
		newMsgChan := make(chan Message)
		return Client{
			Conn: conn, 
			Name: username, 
			MailBoxChan: newMsgChan, 
			}, nil
	}
}

func (h Hub) CleanUpHub(){
	for _, client := range h.clientMap{
		client.Conn.Close()
	}

	if h.joinChannel != nil {
		close(h.joinChannel)
	}
	if h.leaveChannel != nil {
		close(h.leaveChannel)
	}

	os.Exit(0)
}

