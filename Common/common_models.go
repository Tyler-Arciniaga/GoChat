package common

import (
	"fmt"
)

type MessageType int

const (
	Broadcast MessageType = iota
	Whisper
	UserData
	ServerMessage
	Join
	Leave
	FileMetaData //announces new file metadata
	FileData     //carries file's raw bytes
	Ack
)

type Status int

const (
	Ready Status = iota
	Failed
)

type Message struct {
	Type     MessageType `json:"message_type"`
	From     string      `json:"name"`
	To       string      `json:"to"` //empty string unless /whisper command evoked
	Msg      string      `json:"msg"`
	FileMeta FileHeader  `json:"file_meta"`
	FileData []byte      `json:"data"`
	Status   Status      `json:"status"`
}

func (m Message) String() string {
	return fmt.Sprintf("[%s]: %s", m.From, m.Msg)
}

type FileHeader struct {
	Filename string `json:"filename"`
	FileSize int64  `json:"filesize"`
}

type DisconnectionType int

const (
	Interruption DisconnectionType = iota
	Graceful
)
