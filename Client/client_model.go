package main

import (
	"net"
)

type Client struct {
	name        string
	conn        net.Conn
	MailBoxChan chan []byte
	ErrorChan   chan error
}
