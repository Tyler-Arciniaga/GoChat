package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-chat/common"
	"log/slog"
	"net"
	"os"
)

//var once sync.Once

func (c Client) StartClient() {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		fmt.Println("error connecting to chat server:", err)
		return
	}
	c.conn = conn
	fmt.Println("connection to tcp server set!")
	defer conn.Close()

	go c.HandleIncomingMessages()
	c.SendMessages()
}

func (c Client) HandleIncomingMessages() {
	buf := make([]byte, 1024)
	for {
		_, err := c.conn.Read(buf)
		if err != nil {
			slog.Error("client conn read error", "err", err)
		}

		fmt.Println("message from hub:", string(buf))
	}
}

func (c Client) SendMessages() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
		for scanner.Scan() {
			line := scanner.Bytes()
			newMessage := common.Message{Type: common.Broadcast, Name: c.name, Msg: string(line)}
			marshalledMsg, _ := json.Marshal(newMessage)
			c.conn.Write(marshalledMsg)
		}
	}
}

// func (c Client) RecieveMessages() {
// 	for m := range c.MailBoxChan {
// 		if m.Type == Leave {
// 			return
// 		}

// 		var msgString string
// 		if m.Name == "" {
// 			msgString = m.Msg
// 		} else {
// 			msgString = m.String()
// 		}

// 		_, err := c.Conn.Write([]byte(msgString))
// 		if err != nil {
// 			fmt.Printf("client with name %s had trouble printing incoming message, err: %s\n", c.Name, err)
// 			once.Do(func() {
// 				c.HandleDisconnect(leaveChannel, LeaveSignal{LeaveType: Interruption, Client: c})
// 			})
// 			return
// 		}
// 	}
// }

// func (c Client) SendMessages() {
// 	bufferedReader := bufio.NewReader(c.Conn)
// 	for {
// 		bytes, err := bufferedReader.ReadBytes(byte('\n')) //function call blocking until delimeter seen
// 		if err != nil {
// 			if err != io.EOF {
// 				fmt.Printf("client with name %s had trouble sending a message: %s\n", c.Name, err)
// 				once.Do(func() {
// 					c.HandleDisconnect(leaveChannel, LeaveSignal{LeaveType: Interruption, Client: c})
// 				})
// 				return
// 			}
// 		}
// 		msgString := strings.TrimSpace(string(bytes))
// 		if len(msgString) == 0 {
// 			continue
// 		}

// 		if string(msgString[0]) == "/" {
// 			parsedCommand := strings.Split(msgString[1:], " ")
// 			//handle if parsed command len == 0
// 			cmd := parsedCommand[0]

// 			switch cmd {
// 			case "leave":
// 				c.HandleDisconnect(leaveChannel, LeaveSignal{LeaveType: Graceful, Client: c})
// 				return
// 			case "whisper":
// 				go func() {
// 					broadcastChannel <- Message{Type: Whisper, Name: c.Name, To: parsedCommand[1], Msg: strings.Join(parsedCommand[2:], " ")}
// 				}()
// 				continue
// 			case "sendfile":
// 				fmt.Println("sending file...")
// 			default:
// 				continue
// 			}
// 		}

// 		newMessage := Message{Type: Broadcast, Name: c.Name, Msg: msgString}

// 		broadcastChannel <- newMessage
// 	}
// }

// func (c Client) HandleDisconnect(l chan LeaveSignal, s LeaveSignal) {
// 	l <- s
// }
