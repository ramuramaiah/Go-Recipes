package main

import (
	"context"
	"log"
	"net"
	"strings"

	pb "Go-Recipes/grpc-greeting/greeting"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedGreetingServiceServer
}

func (*server) Greeting(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received Greeting from: %v", in.GetName())
	var hobbiesAsStr strings.Builder
	hobbies := in.GetHobbies()
	for i, hobby := range hobbies {
		hobbiesAsStr.WriteString(hobby)
		if i < len(hobbies)-1 {
			hobbiesAsStr.WriteString(",")
		}
	}

	return &pb.HelloResponse{Greeting: "Hello " + in.GetName() + ". " + "Glad to know, your hobbies are: " + hobbiesAsStr.String()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreetingServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
