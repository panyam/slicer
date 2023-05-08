package harness

import (
	"context"
	"fmt"
	"github.com/panyam/slicer/clients"
	"github.com/panyam/slicer/cmd/echosvc"
	"github.com/panyam/slicer/protos"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"time"
)

type Producer struct {
	Addr        string
	ControlAddr string
	Prefix      string
	grpcServer  *grpc.Server
	Logger      *log.Logger
}

func NewProducer(prefix string, addr string, control_addr string, logfile io.Writer) *Producer {
	out := Producer{
		Prefix:      prefix,
		Addr:        addr,
		ControlAddr: control_addr,
		grpcServer:  grpc.NewServer(
		// grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		// grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		),
		Logger: log.New(logfile, fmt.Sprintf("PROD[%s] ", addr), log.Ldate|log.Ltime|log.Lshortfile),
	}
	return &out
}

func (p *Producer) Start() {
	clientMgr := clients.NewStaticClientMgr(p.ControlAddr, protos.NewControlServiceClient)
	ctrlSvcClient, err := clientMgr.GetClient("")
	p.Logger.Println("CC, E: ", ctrlSvcClient, err)
	if err != nil {
		panic(err)
	}

	// do the ping
	go func() {
		t := time.NewTicker(time.Second * 5)
		for {
			<-t.C
			ctrlSvcClient.Client.PingTarget(context.Background(), &protos.PingTargetRequest{
				Address: p.Addr,
			})
		}
	}()

	echosvc.RegisterEchoServiceServer(p.grpcServer, echosvc.NewEchoService(p.Prefix))
	// grpc_prometheus.Register(grpcServer)
	p.Logger.Printf("Initializing Echo Server on %s", p.Addr)
	lis, err := net.Listen("tcp", p.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	p.grpcServer.Serve(lis)
}

func (p *Producer) Stop() {
	p.grpcServer.Stop()
}
