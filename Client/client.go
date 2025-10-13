package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	common "go-chat/Common"
	"io"
	"log/slog"
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
	var e common.Envelope
	for b := range c.MailBoxChan {
		err := json.Unmarshal(b, &e)
		if err != nil {
			// slog.Error("error unmarshalling incoming message", "error", err)
			c.ErrorChan <- err
			return
		}
		switch e.Type {
		case common.Ack:
			var a common.Acknowledgement
			json.Unmarshal(b, &a)
			go func() {
				c.AckChan <- a.Status
			}()
		case common.FileMetaData:
			var f common.FileHeader
			json.Unmarshal(b, &f)
			go func() {
				c.HandleIncomingFileHeader(f)
			}()
		case common.FileData:
			fmt.Println("recieved some file data")
			var d common.FileDataChunk
			json.Unmarshal(b, &d)
			c.FileDataChan <- d
		default:
			var m common.Message
			json.Unmarshal(b, &m)
			fmt.Println(m)
		}

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

		var newMessage any
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) < 1 {
				continue
			}
			if line[0] == byte('/') {
				newMessage, err = c.ParseCommandMessage(string(line)) //TODO: make interface for all client messages (chats, leave signals, file transfers, etc)
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

			byteLength := int32(len(marshalledMsg))
			binary.Write(c.conn, binary.BigEndian, byteLength)
			c.conn.Write(marshalledMsg)
			_, ok := newMessage.(common.LeaveSignal)
			if ok {
				c.ErrorChan <- errors.New("leave signal bundled as error")
			}
		}
	}
}

func (c Client) ParseCommandMessage(line string) (any, error) {
	parsedCommand := strings.Split(line[1:], " ")
	if len(parsedCommand) < 1 {
		return nil, errors.New("error parsing command")
	}

	cmd := parsedCommand[0]
	switch cmd {
	case "leave":
		return common.LeaveSignal{Type: common.Leave, From: c.name, Conn: c.conn}, nil
	case "whisper":
		return common.Message{Type: common.Whisper, From: c.name, To: parsedCommand[1], Msg: strings.Join(parsedCommand[2:], " ")}, nil
	case "sendfile":
		err := c.HandleFileTransfer(parsedCommand[1])
		if err != nil {
			slog.Error(err.Error())
		}
	default:
		return nil, errors.New("error parsing client command line: invalid command")
	}
	return nil, errors.New("error parsing client command line: invalid command")
}

func (c Client) HandleFileTransfer(filename string) error {
	err := c.HandleSendFileHeader(filename)
	if err != nil {
		return fmt.Errorf("error sending file header to server: %s", err)
	}

	ackStatus := <-c.AckChan
	if ackStatus != common.Ready {
		return fmt.Errorf("error: server could not prepare for incoming file data: %s", err)
	}

	err = c.SendFileData(filename)
	if err != nil {
		return fmt.Errorf("error sending file data stream to hub: %s", err)
	}

	return nil
}

func (c Client) SendFileData(filename string) error {
	file, _ := os.Open(filename)
	defer file.Close()

	// _, err := io.Copy(c.conn, file)
	// if err != nil {
	// 	return err
	// }
	var chunk_size int64 //TODO experiment with different chunk size (maybe try speed with go testing package)
	var buf bytes.Buffer

	chunk_size = 1000 //1000 bytes
	for {
		_, err := io.CopyN(file, &buf, chunk_size)
		if err != nil && err != io.EOF {
			return err
		}
		newDataChunk := common.FileDataChunk{Type: common.FileData, From: c.name, DataChunk: buf, IsLast: err == io.EOF}
		b, err := json.Marshal(newDataChunk)
		if err != nil {
			return err
		}

		c.conn.Write(b)
		if err == io.EOF {
			break
		}
	}

	// //fileData := common.FileDataStream{Type: common.FileData, From: c.name, Data: buf}
	// b, err := json.Marshal(fileData)
	// if err != nil {
	// 	return err
	// }
	//
	// c.conn.Write(b)
	return nil
}

func (c Client) HandleSendFileHeader(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileHeader, err := c.ExtractFileMetaData(file)
	if err != nil {
		return err
	}

	b, _ := json.Marshal(fileHeader)
	_, err = c.conn.Write(b)

	if err != nil {
		return err
	}

	return nil
}

func (c Client) HandleIncomingFileHeader(f common.FileHeader) {
	fmt.Printf("User (%s) is trying to send you a file. Accept and download file? Y/N\n", f.From)
	ack, err := c.HandleIncomingFileChoice()
	if err != nil || ack.Status != common.Ready {
		return
	}

	//TODO: send ack to room which gives it to server which sends to client so that it knows it can send now

	go c.HandleIncomingFileData(f)

}

func (c Client) HandleIncomingFileData(f common.FileHeader) {
	filename := f.Filename
	newFile, err := os.Create(fmt.Sprintf("client-downloads/(%s)%s", c.name, filename))
	if err != nil {
		slog.Error("error creating new file based on incoming file header", "err", err)
		return
	}
	defer newFile.Close()

	for chunk := range c.FileDataChan {
		_, err := io.Copy(newFile, &chunk.DataChunk)
		if err != nil {
			slog.Error("error writing from incoming data chunk to local file copy", "err", err)
			return
		}
		if chunk.IsLast {
			break
		}
	}

}

func (c Client) HandleIncomingFileChoice() (common.Acknowledgement, error) {
	//TODO: handle logic regarding giving user choice to donwload or refuse sent file
	return common.Acknowledgement{Type: common.Ack, Status: common.Ready}, nil
}

func (c Client) ExtractFileMetaData(file *os.File) (common.FileHeader, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return common.FileHeader{}, nil
	}

	newFileHeader := common.FileHeader{
		Type:     common.FileMetaData,
		From:     c.name,
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
