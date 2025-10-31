package common

import (
	"fmt"
	"net"
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
	Type MessageType `json:"type"`
	From string      `json:"name"`
	To   string      `json:"to"` //empty string unless /whisper command evoked
	Msg  string      `json:"msg"`
}

func (m Message) String() string {
	return fmt.Sprintf("[%s]: %s", m.From, m.Msg)
}

type LeaveSignal struct {
	Type MessageType `json:"type"`
	From string      `json:"from"`
	Conn net.Conn    `json:"conn"`
}

type Acknowledgement struct {
	Type   MessageType `json:"type"`
	Status Status      `json:"status"`
}

type FileHeader struct {
	Type     MessageType `json:"type"`
	From     string      `json:"from"`
	Filename string      `json:"filename"`
	FileSize int64       `json:"filesize"`
}

type FileDataChunk struct {
	Type      MessageType `json:"type"`
	ChunkNum  int         `json:"chunk_num"`
	From      string      `json:"from"`
	DataChunk []byte      `json:"data_chunk"`
	IsLast    bool        `json:"is_last"`
}

type Envelope struct {
	Type MessageType `json:"type"`
}

type DisconnectionType int

const (
	Interruption DisconnectionType = iota
	Graceful
)
