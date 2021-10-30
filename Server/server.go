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
	wait := &sync.WaitGroup{} // waits for the go routines to finish
	done := make(chan int)    // to know when all the go routines are finished

	for _, c := range s.Connection {
		wait.Add(1) // add new go routine to wait group

		go func(message *proto.Message, c *Connection) {
			defer wait.Done()

			if c.active {
				err := c.stream.Send(message) // send message back to the client that is attached connection
				log.Println("Message being sent to: " + c.id)

				if err != nil {
					log.Println("Could not send message to: " + c.id)
					c.active = false
				}
			}

		}(message, c)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done // block the return statement until routines are done. Done needs to return something before we can return something

	donetwo := make(chan int)

	for _, c := range s.Connection {
		//log.Print("Client is ", c.active)
		if !c.active {
			go sendLeaveMessage(c.id, wait, s.Connection)
		//	log.Print("It is false")
		}
	}

	//log.Print("done two func")

	go func() {
		wait.Wait()
		close(donetwo)
	}()

	<-donetwo // block the return statement until routines are done. Done needs to return something before we can return something
	//log.Println("Done with two")
	return &proto.Close{}, nil
}

func sendLeaveMessage(id string, wait *sync.WaitGroup, conncections []*Connection) {

	for _, c := range conncections {
		wait.Add(1) // add new go routine to wait group

		go func(c *Connection) {
			defer wait.Done()

			if c.active {
				message := &proto.Message{
					Id:        id,
					Message:   "I left the chat",
					Timestamp: 1, // change
				}

				err := c.stream.Send(message) // send message back to the client that is attached connection
				log.Println("Message being sent to: " + c.id)

				if err != nil {
					log.Println("Could not send message to: " + c.id)
					c.active = false
				}
			}

		}(c)
	}
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
	serveError := serverGrpc.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}
