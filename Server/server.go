package main

import(
 "log"
// "context"
 "net"

"google.golang.org/grpc"

)

const(
    port = ":8000"
)

type server struct{
    ChannelOut chan string
}

func main(){
    lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
    if err := s.Serve(lis); err != nil {
    log.Fatalf("failed to serve: %v", err)
    }
}

/* 
func (s *server) Join (ctx context.Context, in *pb.JoinRequest) (*pb.JoinConfirm, error){
}

func (s *server) Publish(){
}

func (s *server) Broadcast(){
} */
