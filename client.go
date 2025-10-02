package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

var once sync.Once

func (c Client) RecieveMessages(){
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
				c.DisconnectFromHub(c.ActiveLeaveChan)
			})
			return
		}
	}
}


func (c Client) SendMessages(){
	fmt.Println("herexxx")
	bufferedReader := bufio.NewReader(c.Conn)
	for {
		bytes, err := bufferedReader.ReadBytes(byte('\n')) //function call blocking until delimeter seen
		if err != nil {
			if err != io.EOF{
				fmt.Printf("client with name %s had trouble sending a message: %s\n", c.Name, err)
				once.Do(func(){
					c.DisconnectFromHub(c.ActiveLeaveChan)
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
				c.DisconnectFromHub(c.ActiveLeaveChan)
				return
			case "whisper":
				go func(){
					c.ActiveRoomChan <- Message{Type: Whisper, Name: c.Name, To: parsedCommand[1], Msg: strings.Join(parsedCommand[2:], " ")}
				}()
				continue
			default:
				continue
			}	
		}


		newMessage := Message{Type: Broadcast, Name: c.Name, Msg: msgString}
		fmt.Println("xyxxyxyx")
		fmt.Println(c.ActiveRoomChan)
		c.ActiveRoomChan <- newMessage
		
	}
}

//TODO: update so that client only leave current chat room not entire chat server (can rejoin a different room, etc)
func (c Client) DisconnectFromHub(leaveChannel chan Client){
	c.Conn.Close()
	close(c.MailBoxChan)
	leaveChannel <- c
}