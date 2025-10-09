package main

import (
	"encoding/json"
	"fmt"
	common "go-chat/Common"
)

func (r Room) StartRoom() {
	go r.RouteRawMessages()
	go r.HandleAdminSignals()
	go r.HandleChatMessages()
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

		case common.FileMetaData:
			fmt.Println("handle file meta data...") //TODO

		case common.FileData:
			fmt.Println("handle file byte data...") //TODO
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

func (r Room) RemoveUser(s ClientModel) {} //TODO

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

func (r Room) BroadcastMessage(m common.Message) {
	b, _ := json.Marshal(m)
	for _, conn := range r.chatterMap {
		go func() {
			conn.Write(b)
		}()
	}
} //TODO

func (r Room) WhisperMessage(m common.Message) {} //TODO

// func (r Room) HandleConnectionMap(hubLeaveChannel chan Client) {
// 	for {
// 		select {
// 		case m := <-r.messageChannel:
// 			if m.Type == Broadcast {
// 				r.BroadCastMessage(m)
// 			} else {
// 				r.WhisperMessage(m)
// 			}

// 		case j := <-r.joinChannel:
// 			r.connectionMap[strings.ToLower(j.Name)] = *j
// 			go func() {
// 				r.messageChannel <- Message{Type: Broadcast, Name: fmt.Sprintf("Room %d", r.roomID), Msg: fmt.Sprintf("%s has joined this chat room", j.Name)}
// 			}()

// 		case l := <-r.leaveChannel:
// 			delete(r.connectionMap, strings.ToLower(l.Client.Name))

// 			if l.LeaveType == Graceful {
// 				go func() {
// 					l.Client.MailBoxChan <- Message{Type: Leave}
// 				}()
// 			}

// 			go func() {
// 				r.messageChannel <- Message{Type: Broadcast, Name: "", Msg: fmt.Sprintf("%s has disconnected from this chat room\n", l.Client.Name)}
// 			}()

// 			if l.LeaveType == Interruption {
// 				go func() {
// 					hubLeaveChannel <- l.Client
// 				}()
// 			}
// 		}
// 	}
// }

// func (r Room) BroadCastMessage(m Message) {
// 	for _, client := range r.connectionMap {
// 		go func() {
// 			if m.Name == client.Name {
// 				newMsg := m
// 				newMsg.Name = "You"
// 				client.MailBoxChan <- newMsg
// 			} else {
// 				client.MailBoxChan <- m
// 			}
// 		}()
// 	}
// }

// func (r Room) WhisperMessage(m Message) {
// 	//TODO: fix issue regarding trying to whisper to users with a multi-word name (not sure how to know when to divide name with message)
// 	dstClient, ok := r.connectionMap[strings.ToLower(m.To)]
// 	srcClient := r.connectionMap[m.Name]
// 	if !ok {
// 		go func() {
// 			srcClient.MailBoxChan <- Message{Name: "", Msg: fmt.Sprintf("Whisper message to: %s failed, perhaps check spelling\n", m.To)}
// 		}()
// 		return
// 	}
// 	go func() {
// 		dstClient.MailBoxChan <- m
// 	}()

// 	go func() {
// 		newMessage := m
// 		newMessage.Name = "You"
// 		srcClient.MailBoxChan <- newMessage
// 	}()
// }
