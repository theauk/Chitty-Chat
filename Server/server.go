package main

import(
"log"
"context"
"net"

"google.golang.org/grpc"
//pb "chittychat"
)

const(
    port = ":8008"
)

type server struct{
 //pb.CreateTheServerYay
 
 ChannelOut chan string
}

func main(){
    lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
    log.Printf("server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
    log.Fatalf("failed to serve: %v", err)
    }
}

func (s *server) Join (ctx context.Context, in *pb.JoinRequest) (*pb.JoinConfirm, error){
}

func (s *server) Publish(){
}

func (s *server) Broadcast(){
}