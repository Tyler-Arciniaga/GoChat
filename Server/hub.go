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
		}
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
			var d common.FileDataChunk
			json.Unmarshal(buf, &d)
			h.HandleIncomingFileStream(conn, d, room)
			if d.IsLast == true {
				go h.RecieveClientMessages(conn)
				return
			}
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
