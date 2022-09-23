package main

import (
	"flag"
	"github.com/panyam/slicer/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	addr = flag.String("addr", "localhost:7000", "Address to run the echo service on.")
)

func main() {
	flag.Parse()
	grpcServer := grpc.NewServer(
	// grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	// grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	service.RegisterControlServiceServer(grpcServer, service.NewControlService())
	// grpc_prometheus.Register(grpcServer)
	log.Printf("Initializing Control Server on %s", *addr)
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer.Serve(lis)
}
