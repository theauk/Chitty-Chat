package main

import (
	//"context"
	//"log"

	"context"
	"fmt"
	"sync"

	//første er init navnet
	proto "chittychat/ChittyChat"
)

var client proto.BroadcastClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *proto.Client) error {
	var streamError error

	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("Failed connection")
	}

	wait.Add(1)
	go func(str proto.Broadcast_CreateStreamClient) { // the streaming part of the client

		defer wait.Done()

		for {

			msg, err := str.Recv() // wait for us to recieve a message from the server
			if err != nil {
				streamError = fmt.Errorf("Could not read message")
				break
			}

			fmt.Printf(msg.Id, msg.Message) // log ?
		}

	}(stream)

}

func main() {

}
