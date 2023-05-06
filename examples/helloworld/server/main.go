package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/vizee/gapi/examples/helloworld/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Greeter struct {
	proto.UnimplementedGreeterServer
}

func (*Greeter) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Printf("SayHello: %s", req)

	if len(req.Name) < 3 {
		return nil, status.Errorf(1000, "name too short")
	}

	return &proto.HelloReply{
		Message: fmt.Sprintf("Hello %s(%s)", req.Name, req.UserId),
	}, nil
}

func main() {
	const listenAddr = ":50051"

	srv := grpc.NewServer()
	proto.RegisterGreeterServer(srv, &Greeter{})
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("listening: %s", listenAddr)

	err = srv.Serve(ln)
	if err != nil {
		log.Fatalf("serve: %v", err)
	}
}
