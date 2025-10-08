package common

import (
	"fmt"
)

type MessageType int

const (
	Broadcast MessageType = iota
	Whisper
	Join
	Leave
	FileMetaData //announces new file metadata
	FileData     //carries file's raw bytes
)

type Message struct {
	Type     MessageType `json:"message_type"`
	Name     string      `json:"name"`
	To       string      `json:"to"` //empty string unless /whisper command evoked
	Msg      string      `json:"msg"`
	FileMeta *FileHeader `json:"file_meta"`
	Data     []byte      `json:"data"`
}

func (m Message) String() string {
	return fmt.Sprintf("[%s]: %s\n", m.Name, m.Msg)
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
