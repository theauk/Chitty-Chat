package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	proto "chittychat/ChittyChat"

	"google.golang.org/grpc"
)

// Global variables
var client proto.BroadcastClient
var wait *sync.WaitGroup
var clientTime int64 = 0

// The wait group will be synced as the first step
func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *proto.Client) error {
	var streamError error

	// Create a new client
	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("failed connection")
	}

	wait.Add(1)
	go func(str proto.Broadcast_CreateStreamClient) { // The streaming part of the client for messages
		defer wait.Done()
		for {
			msg, err := str.Recv() // Wait until receiving a message from the server
			if err != nil {
				streamError = fmt.Errorf("could not read message")
				break
			}
			updateTime(msg.Timestamp)
			log.Printf("Lamport timestamp: %d | Client %s says: %s", clientTime, msg.Id, msg.Message)
		}

	}(stream)

	return streamError
}

func updateTime(incomingTime int64) {
	if incomingTime > clientTime {
		clientTime = incomingTime
	}
	clientTime++
}

func main() {
	done := make(chan int)

	// Get client id from command line
	id := os.Args[1]
	waiter := &sync.WaitGroup{}

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Client could not connect to the given target")
	}

	// Create new gRPC client
	client = proto.NewBroadcastClient(conn)
	chatter := &proto.Client{
		Id: id,
	}

	connectError := connect(chatter)
	if connectError != nil {
		log.Fatalf("Client could not connect")
	}

	sendJoinMessage(chatter)
	sendClientMessages(chatter)

	go func() { // Wait for our wait group decrementing
		waiter.Wait()
		close(done)
	}()

	<-done // Wait until done sends back some data
}

func sendJoinMessage(chatter *proto.Client) {
	clientTime++
	newChatterMsg := &proto.Message{
		Id:        chatter.Id,
		Message:   "I'm joining",
		Timestamp: clientTime,
	}
	_, joinErr := client.BroadcastMessage(context.Background(), newChatterMsg)

	if joinErr != nil {
		log.Println("error sending join message: ", joinErr)
	}
}

func sendClientMessages(chatter *proto.Client) {
	scanner := bufio.NewScanner(os.Stdin) // Scan the input from the user through the command line
	for scanner.Scan() {
		clientTime++
		message := &proto.Message{
			Id:        chatter.Id,
			Message:   scanner.Text(),
			Timestamp: clientTime,
		}
		_, err := client.BroadcastMessage(context.Background(), message)

		if err != nil {
			log.Println("error sending message: ", err)
			break
		}
	}
}
