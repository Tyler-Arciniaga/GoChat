package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

var once sync.Once

func (c Client) RecieveMessages(leaveChannel chan Client){
	for m := range c.MailBoxChan{
		var msgString string
		if m.Name == ""{
			msgString = m.Msg
		} else {
			msgString = m.String()
		}

		_, err := c.Conn.Write([]byte(msgString))
		if err != nil {
			fmt.Printf("client with name %s had trouble printing incoming message: %s\n", c.Name, err)
			once.Do(func(){
				c.DisconnectFromHub(leaveChannel)
			})
			return
		}
	}
}


func (c Client) SendMessages(broadcastChannel chan Message, leaveChannel chan Client){
	bufferedReader := bufio.NewReader(c.Conn)
	for {
		bytes, err := bufferedReader.ReadBytes(byte('\n')) //function call blocking until delimeter seen
		if err != nil {
			if err != io.EOF{
				fmt.Printf("client with name %s had trouble sending a message: %s\n", c.Name, err)
				once.Do(func(){
					c.DisconnectFromHub(leaveChannel)
				})
				return
			}
		}
		msgString := strings.TrimSpace(string(bytes))
		if len(msgString) == 0{
			continue
		}

		if string(msgString[0]) == "/"{
			parsedCommand := strings.Split(msgString[1:], " ")
			//handle if parsed command len == 0
			cmd := parsedCommand[0]

			switch cmd{
			case "leave":
				c.DisconnectFromHub(leaveChannel)
				return
			case "whisper":
				go func(){
					broadcastChannel <- Message{Type: Whisper, Name: c.Name, To: parsedCommand[1], Msg: strings.Join(parsedCommand[2:], " ")}
				}()
				continue
			default:
				continue
			}	
		}


		newMessage := Message{Type: Broadcast, Name: c.Name, Msg: msgString}
		
		broadcastChannel <- newMessage //TODO: writing to nil chan when a user first enters a room
		fmt.Println("made it past")
	}
}

//TODO: update so that client only leave current chat room not entire chat server (can rejoin a different room, etc)
func (c Client) DisconnectFromHub(leaveChannel chan Client){
	c.Conn.Close()
	close(c.MailBoxChan)
	leaveChannel <- c
}