package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main(){
	if len(os.Args) < 2{
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}
	port := fmt.Sprintf(":%s", os.Args[1])

	ln, err := net.Listen("tcp", port)
	if err != nil{
		fmt.Println("error creating network listener")
		os.Exit(1)
	}

	fmt.Println("Listening on port", port)
	for {
		conn, err := ln.Accept()
		if err != nil{
			fmt.Println("error accepting new user connection", err)
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()

	bufReader := bufio.NewReader(conn)
	for {
		bytes, err := bufReader.ReadBytes(byte('\n'))
		if err != nil{
			if err != io.EOF{
				fmt.Println("error reading in bytes from client", err)
			}
			return //triggers conn.Close()
		}

		strReq := string(bytes)
		if strReq == "/leave\n"{
			break
		}
		fmt.Print("request: ", strReq)

		res := fmt.Sprint("you said: ", strReq)
		fmt.Printf("server response: %s", res)

		_, err = conn.Write([]byte(res))
		if err != nil{
			fmt.Println("error writing response back through conn:", err)
			return
		}
	}

}