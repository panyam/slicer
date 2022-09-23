package clients

import (
	// "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"log"
)

type RpcClient[T any] struct {
	// Note that Addr can be uniquely used to identify a client
	Addr   string
	Conn   *grpc.ClientConn
	Client T
}

func (rc *RpcClient[T]) Close() {
	rc.Conn.Close()
}

func NewRpcClient[T any](addr string, ClientCreator func(grpc.ClientConnInterface) T) (*RpcClient[T], error) {
	// var opts []grpc.DialOption
	// conn, err := grpc.Dial(addr, opts...)
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
	// grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),

	if err != nil {
		log.Println("Error creating client on address: ", addr, err)
		return nil, err
	}
	out := &RpcClient[T]{
		Addr:   addr,
		Conn:   conn,
		Client: ClientCreator(conn),
	}
	return out, err
}

type ClientMgr[T any] interface {
	GetClient(entityId string) (*RpcClient[T], error)
}
