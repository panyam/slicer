package main

import (
	"flag"
	"github.com/panyam/slicer/cmd/echosvc"
	"google.golang.org/grpc"
	"log"
	"net"
)

/**
 * A simple binary that acts as a shard provider.
 * Shard providers are services that are running performing a simple function on its input.
 * They append the provider prefix to the input (which is the shard key).
 * The providers are only allowed to serve requests
 * if they own a particular shard.
 */

var (
	addr   = flag.String("addr", "localhost:9000", "Address to run the echo service on.")
	prefix = flag.String("prefix", "white", "The color prefix this provider will prepend requests with")
)

func main() {
	flag.Parse()
	grpcServer := grpc.NewServer(
	// grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	// grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	echosvc.RegisterEchoServiceServer(grpcServer, echosvc.NewEchoService(*prefix))
	// grpc_prometheus.Register(grpcServer)
	log.Printf("Initializing Echo Server on %s", *addr)
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer.Serve(lis)
}
