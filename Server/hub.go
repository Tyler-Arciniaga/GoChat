package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	common "go-chat/Common"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
)

func (h Hub) Start() {
	defer func() {
		if v := recover(); v != nil {
			slog.Error("Unrecoverable panic detected, shutting down chat server now.", "panic", v)
			h.CleanUpHub()
		}
	}()
	defer h.CleanUpHub()

	slog.Info(fmt.Sprintf("Listening on port %s", h.port))

	ln, err := net.Listen("tcp", h.port)
	if err != nil {
		fmt.Println("error creating network listener")
		os.Exit(1)
	}

	go h.HandleClientDisconnect()

	for i := range 3 {
		newRoom := Room{
			roomID:            i,
			chatterMap:        make(map[string]net.Conn),
			messageChannel:    make(chan common.Message),
			joinChannel:       make(chan ClientModel),
			leaveChannel:      make(chan ClientModel),
			broadcastChannel:  make(chan common.Message),
			whisperChannel:    make(chan common.Message),
			fileHeaderChannel: make(chan common.FileHeader),
			fileDataChannel:   make(chan common.FileDataChunk),
		}

		h.roomMap[i] = &newRoom

		newRoom.StartRoom()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting new user connection", err)
			continue
		}
		slog.Info("new client connection established", "conn", conn)

		go h.HandleNewClientConnection(conn)

	}
}

func (h Hub) HandleNewClientConnection(conn net.Conn) {
	clientName, err := h.HandleClientInfoMessage(conn) //blocking call
	if err != nil {
		h.leaveChannel <- conn
		return
	}

	room, err := h.HandleRoomSelect(conn)
	if err != nil {
		slog.Error("error selecting room", "error", err)
		h.leaveChannel <- conn
		return
	}

	joinSignal := ClientModel{name: clientName, conn: conn}
	go func() {
		room.joinChannel <- joinSignal
	}()

	h.clientRoomMap[conn] = room

	go h.RecieveClientMessages(conn) //start recieving client messages
}

func (h Hub) HandleClientInfoMessage(conn net.Conn) (string, error) {
	buf := make([]byte, 1024)
	var msg common.Message
	for {
		n, err := conn.Read(buf)
		if err != nil {
			slog.Error("client conn read error", "err", err)
			return "", err
		}

		json.Unmarshal(buf[:n], &msg)
		if msg.Type == common.UserData {
			return msg.From, nil
		} //TODO: handle avoiding duplicate names
	}
}

func (h Hub) RecieveClientMessages(conn net.Conn) {
	var msgLength int32
	var envelope common.Envelope
	room := h.clientRoomMap[conn]
	for {
		if err := binary.Read(conn, binary.BigEndian, &msgLength); err != nil {
			return
		}
		buf := make([]byte, msgLength)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			slog.Error("error reading in client message", "error", err)
			h.leaveChannel <- conn
			return
		}

		err = json.Unmarshal(buf, &envelope)
		if err != nil {
			slog.Error("error unmarshalling incoming message bytes into envelope type", "error", err)
			continue
		}

		switch envelope.Type {
		case common.Leave:
			var l common.LeaveSignal
			json.Unmarshal(buf, &l)
			go func() {
				h.leaveChannel <- conn
			}()
			go func() {
				room.leaveChannel <- ClientModel{name: l.From, conn: conn}
			}()
		case common.FileMetaData:
			var f common.FileHeader
			json.Unmarshal(buf, &f)
			go func() {
				h.HandleIncomingFileHeader(conn, f, room)
			}()
		case common.FileData:
			// fmt.Println("handle file data stream") //TODO
			var d common.FileDataChunk
			json.Unmarshal(buf, &d)
			h.HandleIncomingFileStream(conn, d, room)
		default:
			var m common.Message
			json.Unmarshal(buf, &m)
			fmt.Println(string(buf))
			room.messageChannel <- m
		}
	}
}

func (h Hub) HandleIncomingFileHeader(conn net.Conn, f common.FileHeader, r *Room) error {
	//reroute the file header to the clients in the same room
	r.fileHeaderChannel <- f
	//send ack message
	ackMessage := common.Acknowledgement{Type: common.Ack, Status: common.Ready} //TODO: improve ack logic
	b, _ := json.Marshal(ackMessage)
	h.WriteDirectClientMessage(conn, b)

	// filename := f.Filename
	// filesize := f.FileSize

	// newFile, err := os.Create(fmt.Sprintf("server-downloads/(server)%s", filename))
	// if err != nil {
	// 	return err
	// }
	// defer newFile.Close()

	// _, err = io.CopyN(newFile, conn, filesize)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (h Hub) HandleIncomingFileStream(conn net.Conn, d common.FileDataChunk, r *Room) error {
	r.fileDataChannel <- d
	return nil
}

func (h Hub) HandleClientDisconnect() {
	for c := range h.leaveChannel {
		slog.Info("handling client connection disconnect", "conn", c)
		delete(h.clientRoomMap, c)
		c.Close()
	}
}

func (h Hub) HandleRoomSelect(conn net.Conn) (*Room, error) {
	roomSelectMsg := common.Message{Type: common.ServerMessage, From: "Server", Msg: "Select a chat room 0, 1, 2 (select -1 to leave server)"}
	invalidChoiceMsg := common.Message{Type: common.ServerMessage, From: "Server", Msg: "Invalid choice, try again"}

	roomChoiceBytes, err := json.Marshal(roomSelectMsg)
	invalidChoiceBytes, err2 := json.Marshal(invalidChoiceMsg)

	if err != nil || err2 != nil {
		slog.Error("error marshalling room select message", "error", err)
	}

	h.WriteDirectClientMessage(conn, roomChoiceBytes)
	var msgLength int32
	var msg common.Message
	for {
		if err := binary.Read(conn, binary.BigEndian, &msgLength); err != nil {
			slog.Error("error reading in message length conn string")
		}
		buf := make([]byte, msgLength)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			slog.Error("error reading in client message", "error", err)
		}

		json.Unmarshal(buf, &msg)
		choice, err := strconv.Atoi(msg.Msg)
		if err != nil {
			h.WriteDirectClientMessage(conn, invalidChoiceBytes)
			continue
		}
		if choice == -1 {
			return nil, errors.New("leave command bundled as error")
		}

		room, ok := h.roomMap[choice]
		if !ok {
			conn.Write(invalidChoiceBytes)
			continue
		}

		return room, nil
	}
}

func (h Hub) WriteDirectClientMessage(conn net.Conn, marshalledMsg []byte) error {
	byteLength := int32(len(marshalledMsg))
	err := binary.Write(conn, binary.BigEndian, byteLength)
	if err != nil {
		return err
	}
	conn.Write(marshalledMsg)
	return nil
}

func (h Hub) CleanUpHub() {
	for _, r := range h.roomMap {
		r.CleanUpRoom()
	}
	for c := range h.clientRoomMap {
		c.Close()
	}
	close(h.joinChannel)
	close(h.leaveChannel)
}

// func (h Hub) HandleClientMap() {
// 	for {
// 		select {
// 		case j := <-h.joinChannel:
// 			h.clientMap[strings.ToLower(j.Name)] = j
// 		case l := <-h.leaveChannel:
// 			delete(h.clientMap, strings.ToLower(l.Name))
// 			slog.Info(fmt.Sprintf("Client: %s has disconnected from Go Chat", l.Name))
// 		}
// 	}
// }

// func (h Hub) PromptRoomSelect(c Client) (*Room, error) {
// 	tempReader := bufio.NewReader(c.Conn)
// 	for {
// 		_, err := c.Conn.Write([]byte("Pick a chat room (-1 to disconnect from sever):\n0\n1\n2\n-> ")) //TODO: refactor so that room numbers are not hardcoded
// 		if err != nil {
// 			return nil, fmt.Errorf("error writing to connection: %s", err)
// 		}

// 		bytes, err := tempReader.ReadBytes(byte('\n'))
// 		if err != nil {
// 			return nil, errors.New("error reading in bytes when prompting for room id")
// 		}

// 		roomChoiceStr := strings.TrimSpace(string(bytes))
// 		if roomChoiceStr == "-1" {
// 			c.Conn.Write([]byte("Come back soon :)\n"))
// 			h.leaveChannel <- c
// 			c.Conn.Close()
// 			return nil, nil
// 		}
// 		if len(roomChoiceStr) < 1 {
// 			continue
// 		}
// 		roomChoiceInt, err := strconv.Atoi(roomChoiceStr)
// 		if err != nil {
// 			continue
// 		}

// 		room, ok := h.roomMap[roomChoiceInt]
// 		if !ok {
// 			_, err := c.Conn.Write([]byte("Invalid room number, try again\n"))
// 			if err != nil {
// 				return nil, fmt.Errorf("error writing to connection: %s", err)
// 			}
// 			continue
// 		}
// 		return room, nil
// 	}
// }

// func (h Hub) HandleConnection(conn net.Conn) {
// 	client, err := h.CreateNewClient(conn)
// 	if err != nil {
// 		fmt.Println("error creating new client:", err)
// 		_, err := conn.Write([]byte("error connecting to server try again"))
// 		if err != nil {
// 			slog.Error(fmt.Sprint("error writing error message to connection", err))
// 		}
// 		conn.Close()
// 		return
// 	}

// 	h.joinChannel <- client

// 	for {
// 		room, err := h.PromptRoomSelect(client)
// 		if room == nil {
// 			return
// 		} // user has disconnected from chat server with /leave command
// 		if err != nil {
// 			slog.Error("error with room selection, booting client")
// 			h.leaveChannel <- client
// 			return
// 		}

// 		room.joinChannel <- &client

// 		var wg sync.WaitGroup
// 		wg.Add(2)

// 		go func() {
// 			defer wg.Done()
// 			client.SendMessages(room.messageChannel, room.leaveChannel)
// 		}()

// 		go func() {
// 			defer wg.Done()
// 			client.RecieveMessages(room.leaveChannel)
// 		}()

// 		wg.Wait()
// 	}

// }

// func (h Hub) DisconnectClient(c Client) {
// 	h.leaveChannel <- c
// }

// func (h Hub) CreateNewClient(conn net.Conn) (Client, error) {
// 	tempReader := bufio.NewReader(conn)

// 	for {
// 		_, err := conn.Write([]byte("Enter name before entering chat room: "))
// 		if err != nil {
// 			return Client{}, errors.New("error prompting user for name")
// 		}

// 		bytes, err := tempReader.ReadBytes(byte('\n'))
// 		if err != nil {
// 			if err != io.EOF {
// 				return Client{}, errors.New("error reading in bytes when prompting for name")
// 			}
// 		}

// 		username := strings.TrimSpace(string(bytes))
// 		if len(username) < 1 {
// 			continue
// 		}

// 		_, exists := h.clientMap[strings.ToLower(username)]
// 		if exists {
// 			_, err = conn.Write([]byte("Name already taken in current chat room\n"))
// 			if err != nil {
// 				return Client{}, errors.New("error requesting new name")
// 			}
// 			continue
// 		}

// 		newMsgChan := make(chan Message)
// 		return Client{
// 			Conn:        conn,
// 			Name:        username,
// 			MailBoxChan: newMsgChan,
// 		}, nil
// 	}
// }

// func (h Hub) CleanUpHub() {
// 	for _, client := range h.clientMap {
// 		client.Conn.Close()
// 	}

// 	if h.joinChannel != nil {
// 		close(h.joinChannel)
// 	}
// 	if h.leaveChannel != nil {
// 		close(h.leaveChannel)
// 	}

// 	os.Exit(0)
// }
