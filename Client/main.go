package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <port> <name>")
		os.Exit(1)
	}

	c := Client{name: os.Args[2], MailBoxChan: make(chan []byte), ErrorChan: make(chan error)}
	c.StartClient()
}
