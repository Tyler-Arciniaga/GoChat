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

func (c Client) RecieveMessages(){
	for m := range c.MailBoxChan{
		_, err := c.Conn.Write([]byte(m.String()))
		if err != nil {
			fmt.Printf("client with name %s had trouble printing incoming message: %s", c.Name, err)
			return
		}
	}
}


func (c Client) SendMessages(broadcastChannel chan Message){
	bufferedReader := bufio.NewReader(c.Conn)
	for {
		bytes, err := bufferedReader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF{
				fmt.Printf("client with name %s had trouble sending a message: %s", c.Name, err)
			}
		}

		msgString := strings.TrimSpace(string(bytes))
		newMessage := Message{Name: c.Name, Msg: msgString}
		broadcastChannel <- newMessage
	}
}