package main

import (
	"fmt"
	common "go-chat/Common"
	"os"
)

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <name> <ip-address>")
		fmt.Println("If confused on ip-address run `ip a` (Linux), or `ifconfig` (MacOS)")
		fmt.Println("If ip-address is left blank, program will default to localhost")
		os.Exit(1)
	}

	var ip string
	if len(os.Args) == 3 {
		ip = os.Args[2]
	} else {
		ip = "localhost"
	}

	c := Client{
		name:         os.Args[1],
		ip:           ip,
		MailBoxChan:  make(chan []byte),
		ErrorChan:    make(chan error),
		AckChan:      make(chan common.Status),
		FileDataChan: make(chan common.FileDataChunk),
	}
	c.StartClient()
}
