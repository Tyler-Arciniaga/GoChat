package main

import (
	common "go-chat/Common"
	"net"
)

type Client struct {
	name         string
	conn         net.Conn
	MailBoxChan  chan []byte
	ErrorChan    chan error
	AckChan      chan common.Status
	FileDataChan chan common.FileDataChunk
}
