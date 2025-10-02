package main

import (
	"fmt"
	"strings"
)

func (r Room) HandleConnectionMap(){
	for {
		select{
		case m := <- r.messageChannel:
			if m.Type == Broadcast{
				r.BroadCastMessage(m)
			} else {
				r.WhisperMessage(m)
			}
		case j := <- r.joinChannel:
			r.connectionMap[strings.ToLower(j.Name)] = *j
			go func(){
				r.messageChannel <- Message{Type: Broadcast, Name: fmt.Sprintf("Room %d", r.roomID), Msg: fmt.Sprintf("%s has joined the chat server\n", j.Name)}
			}()
		case l := <- r.leaveChannel:
			delete(r.connectionMap, strings.ToLower(l.Name))
			go func(){
				r.messageChannel <- Message{Type: Broadcast, Name: "", Msg: fmt.Sprintf("%s has disconnected from chat server\n", l.Name)}
			}()
		}
	}
}

func (r Room) BroadCastMessage(m Message){
	for _, client := range r.connectionMap{
		go func(){
			if m.Name == client.Name{
				newMsg := m
				newMsg.Name = "You"
				client.MailBoxChan <- newMsg
			} else{
			client.MailBoxChan <- m
			}
		}()
	}
}

func (r Room) WhisperMessage(m Message){
	//TODO: fix issue regarding trying to whisper to users with a multi-word name (not sure how to know when to divide name with message)
	dstClient, ok := r.connectionMap[strings.ToLower(m.To)]
	srcClient := r.connectionMap[m.Name]
	if !ok{
		go func(){
			srcClient.MailBoxChan <- Message{Name: "", Msg: fmt.Sprintf("Whisper message to: %s failed, perhaps check spelling\n", m.To)}
		}()
		return
	}
	go func(){
		dstClient.MailBoxChan <- m
	}()
	
	go func(){
		newMessage := m
		newMessage.Name = "You"
		srcClient.MailBoxChan <- newMessage
	}()
}