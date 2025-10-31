package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	common "go-chat/Common"
	"net"
)

func (r Room) StartRoom() {
	go r.RouteRawMessages()
	go r.HandleAdminSignals()
	go r.HandleChatMessages()
	go r.HandleFileHeaders()
	go r.HandleFileDataStream()
}

func (r Room) RouteRawMessages() {
	for raw := range r.messageChannel {
		switch raw.Type {
		case common.Broadcast:
			go func() {
				r.broadcastChannel <- raw
			}()

		case common.Whisper:
			go func() {
				r.whisperChannel <- raw
			}()

			// case common.FileMetaData:
			// 	fmt.Println("handle file meta data...") //TODO
			//
			// case common.FileData:
			// 	fmt.Println("handle file byte data...") //TODO
		}
	}
}

func (r Room) HandleAdminSignals() {
	for {
		select {
		case s := <-r.joinChannel:
			go func() {
				r.AdmitUser(s)
			}()
		case s := <-r.leaveChannel:
			go func() {
				r.RemoveUser(s)
			}()
		}
	}
}

func (r Room) AdmitUser(s ClientModel) {
	r.chatterMap[s.name] = s.conn
	r.broadcastChannel <- common.Message{Type: common.Broadcast, From: fmt.Sprintf("Room %d", r.roomID), Msg: fmt.Sprintf("%s has joined the chat room", s.name)}
}

func (r Room) RemoveUser(s ClientModel) {
	disconnectMessage := common.Message{Type: common.Broadcast, From: fmt.Sprintf("Room %d", r.roomID), Msg: fmt.Sprintf("%s has disconnected from the chat room", s.name)}
	go func() {
		r.broadcastChannel <- disconnectMessage
	}()

	delete(r.chatterMap, s.name)
}

func (r Room) HandleChatMessages() {
	for {
		select {
		case m := <-r.broadcastChannel:
			go func() {
				r.BroadcastMessage(m)
			}()

		case m := <-r.whisperChannel:
			go func() {
				r.WhisperMessage(m)
			}()
		}
	}
}

func (r Room) HandleFileHeaders() {
	for f := range r.fileHeaderChannel {
		b, _ := json.Marshal(f)
		for name, conn := range r.chatterMap {
			if name != f.From {
				go func() {
					r.SendLengthPrefixMessage(conn, b)
					conn.Write(b)
				}()
			}
		}
	}
}

func (r Room) HandleFileDataStream() {
	for d := range r.fileDataChannel {
		b, _ := json.Marshal(d)
		for name, conn := range r.chatterMap {
			if name != d.From {
				r.SendLengthPrefixMessage(conn, b)
				conn.Write(b)
			}
		}
	}
}

func (r Room) BroadcastMessage(m common.Message) {
	b, _ := json.Marshal(m)
	for _, conn := range r.chatterMap {
		go func() {
			r.SendLengthPrefixMessage(conn, b)
			conn.Write(b)
		}()
	}
}

func (r Room) WhisperMessage(m common.Message) {
	b, _ := json.Marshal(m)
	dstConn, ok := r.chatterMap[m.To]
	srcConn := r.chatterMap[m.From]
	go func() {
		if !ok {
			serverMessage := common.Message{Type: common.ServerMessage, From: "Server", Msg: "the person you are trying to whisper does not exist"}
			b, _ := json.Marshal(serverMessage)
			r.SendLengthPrefixMessage(srcConn, b)
			srcConn.Write(b)
			return
		}

		r.SendLengthPrefixMessage(dstConn, b)
		dstConn.Write(b)
	}()
}

func (r Room) SendLengthPrefixMessage(conn net.Conn, marshalledMsg []byte) error {
	byteLength := int32(len(marshalledMsg))
	err := binary.Write(conn, binary.BigEndian, byteLength)
	if err != nil {
		return err
	}

	return nil
}

func (r Room) CleanUpRoom() {
	close(r.broadcastChannel)
	close(r.joinChannel)
	close(r.leaveChannel)
	close(r.messageChannel)
	close(r.whisperChannel)
	close(r.fileDataChannel)
}
