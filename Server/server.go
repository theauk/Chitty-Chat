package main

import (
	"context"
	"log"
	"sync"

	// "context"
	proto "chittychat/ChittyChat"
	"net"

	"google.golang.org/grpc"
)

type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	Connection []*Connection // collection of connections (pointing to connections)
}

func (s *Server) CreateStream(pcon *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	con := &Connection{ // ?
		stream: stream,
		id:     pcon.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, con)
	return <-con.error
}

func (s *Server) BroadcastMessage(c context.Context, message *proto.Message) (*proto.Close, error) {
	wait := sync.WaitGroup{} // waits for the go routines to finish
	done := make(chan int)   // to know when all the go routines are finished

	for _, c := range s.Connection {
		wait.Add(1) // add new go routine to wait group

		go func(message *proto.Message, c *Connection) {
			defer wait.Done()

			if c.active {
				err := c.stream.Send(message) // send message back to the client that is attached connection
				log.Println("Message being sent to: ", c.stream)

				if err != nil {
					log.Println("Could not send message")
					c.active = false
					c.error <- err
				}
			}

		}(message, c)
	}

	go func() { // another go routine that runs and ensures that the waitgroup will wait for the other go routines
		wait.Wait()
		close(done)
	}()

	<-done // block the return statement until routines are done. Done needs to return something before we can return something
	return &proto.Close{}, nil
}

func main() {

	var connections []*Connection  // Pointers to connections
	server := &Server{connections} // Make server with connection array

	serverGrpc := grpc.NewServer()              // Start server
	listener, err := net.Listen("tcp", ":8080") // Listen port 8080

	if err != nil {
		log.Fatalf("Could not create the server")
	}

	log.Println("Started server at port 8080")

	proto.RegisterBroadcastServer(serverGrpc, server)
	serverGrpc.Serve(listener)
}
