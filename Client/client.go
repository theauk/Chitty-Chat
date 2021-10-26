package main

import (
	//"context"
	//"log"
	"fmt"
	"log"
	"net"
	"os"
   
    //første er init navnet
	cc "chittychat/ChittyChat"

    "google.golang.org/grpc"
)

const (
	defaultName = "chitty"
)

func main() {
	address := ":" + os.Args[1]
	
    conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := cc.NewChittyClient(conn)
}
