package main

import (
	//"context"
	//"log"
	"fmt"
	"log"
	"net"
	"os"

	cc "Project2/ChittyChat/ChittyChat"
)

const (
	defaultName = "chitty"
)

func main() {
	address := ":" + os.Args[1]
	fmt.Println(address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cc.JoinRequest
}
