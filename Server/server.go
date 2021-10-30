package main

import (
	proto "chittychat/ChittyChat"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

// Server holds a collection of connections (information about the clients)
type Server struct {
	Connection []*Connection
}

// CreateStream implements interface from proto file
func (s *Server) CreateStream(pConnection *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	con := &Connection{
		stream: stream,
		id:     pConnection.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, con)
	return <-con.error
}

// BroadcastMessage implements interface from proto file. The method is used to broadcast messages to all active clients
func (s *Server) BroadcastMessage(c context.Context, message *proto.Message) (*proto.Close, error) {
	// Wait group to keep check of the go routines
	wait := &sync.WaitGroup{}
	done := make(chan int)

	for _, c := range s.Connection {
		wait.Add(1)

		go func(message *proto.Message, c *Connection) {
			defer wait.Done()

			// Only broadcast messages to active clients
			if c.active {
				err := c.stream.Send(message) // Send message back to the client who is attached to the connection
				log.Println("Message being sent to: " + c.id)

				if err != nil {
					log.Println("Could not send message to: " + c.id)
					c.active = false
				}
			}

		}(message, c)
	}

	// Wait until all routines are done
	go func() {
		wait.Wait()
		close(done)
	}()

	<-done // Block the statements below until routines are done

	findNonActiveClients(s, wait)

	return &proto.Close{}, nil
}

func findNonActiveClients(s *Server, wait *sync.WaitGroup) {
	done := make(chan int)

	for _, c := range s.Connection {
		if !c.active {
			go sendLeaveMessage(c.id, wait, s.Connection)
		}
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done // Block the statements below until routines are done
}

func sendLeaveMessage(id string, wait *sync.WaitGroup, connections []*Connection) {

	for _, c := range connections {
		wait.Add(1)

		go func(c *Connection) {
			defer wait.Done()

			if c.active {
				message := &proto.Message{
					Id:      id,
					Message: "I left the chat",
				}

				err := c.stream.Send(message)
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

	proto.RegisterBroadcastServer(serverGrpc, server) // Register the gRPC service
	serveError := serverGrpc.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}
