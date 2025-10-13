package main

import (
	"fmt"
	common "go-chat/Common"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <name>")
		os.Exit(1)
	}

	c := Client{
		name: os.Args[1], 
		MailBoxChan: make(chan []byte), 
		ErrorChan: make(chan error), 
		AckChan: make(chan common.Status),
		FileDataChan: make(chan common.FileDataStream),
	}
	c.StartClient()
}
