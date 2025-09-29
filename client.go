package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type Client struct{
	Conn net.Conn
	Name string
	MailBoxChan chan Message
}

func (c Client) RecieveMessages(leaveChannel chan Client){
	defer c.DisconnectFromHub(leaveChannel)

	for m := range c.MailBoxChan{
		var msgString string
		if m.Name == "Server"{
			msgString = m.Msg
		} else {
			msgString = m.String()
		}

		_, err := c.Conn.Write([]byte(msgString))
		if err != nil {
			fmt.Printf("client with name %s had trouble printing incoming message: %s\n", c.Name, err)
			return
		}
	}
}


func (c Client) SendMessages(broadcastChannel chan Message, leaveChannel chan Client){
	defer c.DisconnectFromHub(leaveChannel)

	bufferedReader := bufio.NewReader(c.Conn)
	for {
		bytes, err := bufferedReader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF{
				fmt.Printf("client with name %s had trouble sending a message: %s\n", c.Name, err)
				return
			}
		}

		msgString := strings.TrimSpace(string(bytes))
		newMessage := Message{Name: c.Name, Msg: msgString}
		broadcastChannel <- newMessage
	}
}

func (c Client) DisconnectFromHub(leaveChannel chan Client){
	leaveChannel <- c
}