package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	common "go-chat/Common"
	"net"
	"os"
	"strings"
)

//var once sync.Once

func (c Client) StartClient() {
	defer func() {
		if v := recover(); v != nil {
			c.CleanUpClient()
		}
	}()
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		fmt.Println("error connecting to chat server:", err)
		return
	}
	c.conn = conn
	fmt.Println("connection to chat server established!")
	defer c.CleanUpClient()

	c.SendClientInfo()

	go c.HandleIncomingMessages()
	go c.PrintIncomingMessages()
	go c.SendMessages()
	c.HandleFatalErrors() //blocking statement, when this returns StartClient finishes and clean up is triggered
}

func (c Client) SendClientInfo() {
	userInfoMsg := common.Message{Type: common.UserData, From: c.name}
	bytes, err := json.Marshal(userInfoMsg)
	if err != nil {
		// slog.Error("error marshalling user info message")
		return
	}

	c.conn.Write(bytes)
}

func (c Client) HandleIncomingMessages() {
	buf := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			// slog.Error("client conn read error", "err", err)
			c.ErrorChan <- err
			return
		}

		c.MailBoxChan <- buf[:n]
	}
}

func (c Client) PrintIncomingMessages() {
	var msg common.Message
	for b := range c.MailBoxChan {
		err := json.Unmarshal(b, &msg)
		if err != nil {
			// slog.Error("error unmarshalling incoming message", "error", err)
			c.ErrorChan <- err
			return
		}

		fmt.Println(msg)
	}
}

func (c Client) SendMessages() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		err := scanner.Err()
		if err != nil {
			// slog.Error("error reading in input from stdin", "error", err)
			c.ErrorChan <- err
		}

		var newMessage common.Message
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) < 1 {
				continue
			}
			if line[0] == byte('/') {
				newMessage, err = c.ParseCommandMessage(string(line))
				if err != nil {
					// slog.Error("error parsing command message from client", "error", err)
					continue
				}
			} else {
				newMessage = common.Message{Type: common.Broadcast, From: c.name, Msg: string(line)}
			}
			marshalledMsg, err := json.Marshal(newMessage)
			if err != nil {
				// slog.Error("error marshaling message data", "error", err)
				continue
			}

			c.conn.Write(marshalledMsg)
			if newMessage.Type == common.Leave {
				c.ErrorChan <- errors.New("leave signal bundled as error")
			}
		}
	}
}

func (c Client) ParseCommandMessage(line string) (common.Message, error) {
	parsedCommand := strings.Split(line[1:], " ")
	if len(parsedCommand) < 1 {
		return common.Message{}, errors.New("error parsing command")
	}

	cmd := parsedCommand[0]
	switch cmd {
	case "leave":
		return common.Message{Type: common.Leave, From: c.name}, nil
	case "whisper":
		return common.Message{Type: common.Whisper, From: c.name, To: parsedCommand[1], Msg: strings.Join(parsedCommand[2:], " ")}, nil
	case "sendfile":
		fmt.Println("handle send file") //TODO
		fmt.Println(os.Getwd())
		err := c.HandleFileTransfer(parsedCommand[1])
		fmt.Println(err)
	default:
		return common.Message{}, errors.New("error parsing client command line: invalid command")
	}
	return common.Message{}, errors.New("error parsing client command line: invalid command")
}

func (c Client) HandleFileTransfer(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileHeader, err := c.ExtractFileMetaData(file)
	if err != nil {
		return err
	}

	fileHeaderMsg := common.Message{Type: common.FileMetaData, From: c.name, FileMeta: fileHeader}
	b, _ := json.Marshal(fileHeaderMsg)
	c.conn.Write(b)

	return nil
}

func (c Client) ExtractFileMetaData(file *os.File) (common.FileHeader, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return common.FileHeader{}, nil
	}

	newFileHeader := common.FileHeader{
		Filename: fileInfo.Name(),
		FileSize: fileInfo.Size(),
	}

	return newFileHeader, nil
}

func (c Client) HandleFatalErrors() {
	for err := range c.ErrorChan {
		fmt.Println("fatal error detected:", err)
		// slog.Error(err.Error())
		return
	}
}

func (c Client) CleanUpClient() {
	fmt.Println("disconnecting...")
	c.conn.Close()
	close(c.MailBoxChan)
	close(c.ErrorChan)
	fmt.Println("see you soon :)")
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
