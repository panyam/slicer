package main

import (
	"context"
	"flag"
	"github.com/panyam/slicer/clients"
	"github.com/panyam/slicer/cmd/echosvc"
	"github.com/panyam/slicer/protos"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

/**
 * A simple binary that acts as a shard provider.
 * Shard providers are services that are running performing a simple function on its input.
 * They append the provider prefix to the input (which is the shard key).
 * The providers are only allowed to serve requests
 * if they own a particular shard.
 */

var (
	addr         = flag.String("addr", "localhost:9000", "Address to run the echo service on.")
	control_addr = flag.String("control_addr", "localhost:7000", "Address where control service is running.")
	prefix       = flag.String("prefix", "white", "The color prefix this provider will prepend requests with")
)

func main() {
	flag.Parse()
	grpcServer := grpc.NewServer(
	// grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	// grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	clientMgr := clients.NewStaticClientMgr(*control_addr, protos.NewControlServiceClient)
	ctrlSvcClient, err := clientMgr.GetClient("")
	log.Println("CC, E: ", ctrlSvcClient, err)
	if err != nil {
		panic(err)
	}

	// do the ping
	go func() {
		t := time.NewTicker(time.Second)
		for {
			<-t.C
			ctrlSvcClient.Client.PingTarget(context.Background(), &protos.PingTargetRequest{
				Address: *addr,
			})
		}
	}()

	echosvc.RegisterEchoServiceServer(grpcServer, echosvc.NewEchoService(*prefix))
	// grpc_prometheus.Register(grpcServer)
	log.Printf("Initializing Echo Server on %s", *addr)
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer.Serve(lis)
}
