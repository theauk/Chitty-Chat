package main

import (
	//"context"
	//"log"

	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	//første er init navnet
	proto "chittychat/ChittyChat"

	"google.golang.org/grpc"
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
		return fmt.Errorf("failed connection")
	}

	wait.Add(1)
	go func(str proto.Broadcast_CreateStreamClient) { // the streaming part of the client

		defer wait.Done()

		for {

			msg, err := str.Recv() // wait for us to recieve a message from the server
			if err != nil {
				streamError = fmt.Errorf("could not read message")
				break
			}
			log.Println("Chatter ", msg.Id, " says: ", msg.Message)
		}

	}(stream)

	return streamError

}

func main() {
	done := make(chan int)

	id := os.Args[1]

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Client could not connect to the service")
	}

	client = proto.NewBroadcastClient(conn)
	chatter := &proto.Client{
		Id: id,
	}

	connect(chatter)

	sendJoinMessage(chatter)

	wait.Add(1) // since we have to create another go routine below

	go func() {
		defer wait.Done() // makes sure we know when wait. finishes

		scanner := bufio.NewScanner(os.Stdin) // to scan the input from the user through the command line
		for scanner.Scan() {
			message := &proto.Message{
				Id:        chatter.Id,
				Message:   scanner.Text(),
				Timestamp: 1, // change
			}
			_, err := client.BroadcastMessage(context.Background(), message)

			if err != nil {
				log.Println("error sending message: ", err)
				break
			}
		}
	}()

	go func() { // Wait for our waitgroup decrementing
		wait.Wait()
		close(done)
	}()

	<-done // Wait until done sends back some data
}

func sendJoinMessage(chatter *proto.Client) {

	newChatterMsg := &proto.Message{
		Id:        chatter.Id,
		Message:   "I'm joining",
		Timestamp: 1, // change
	}

	_, joinErr := client.BroadcastMessage(context.Background(), newChatterMsg)

	if joinErr != nil {
		log.Println("error sending join message: ", joinErr)
	}
}
