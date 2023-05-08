package clients

import (
	"google.golang.org/grpc"
)

// A static client manager
type StaticClientMgr[T any] struct {
	// TODO - turn this into a list?
	StaticAddress string
	CachedClient  *RpcClient[T]
	ClientCreator func(grpc.ClientConnInterface) T
}

func NewStaticClientMgr[T any](addr string, clientCreator func(grpc.ClientConnInterface) T) *StaticClientMgr[T] {
	return &StaticClientMgr[T]{
		StaticAddress: addr,
		ClientCreator: clientCreator,
	}
}

func (ssm *StaticClientMgr[T]) GetClient(entityId string) (_ *RpcClient[T], err error) {
	if ssm.CachedClient != nil {
		ssm.CachedClient, err = NewRpcClient(ssm.StaticAddress, ssm.ClientCreator)
		if err != nil {
			return nil, err
		}
	}
	return ssm.CachedClient, nil
}

// A shardec client manager using the control service
type ShardedClientMgr[T any] struct {
	ClientCreator func(grpc.ClientConnInterface) T

	// Keeps a track of ongoing clients with a given address so we only have 1 connection per address
	Addr2Client map[string]*RpcClient[T]
}

func NewShardedClientMgr[T any](clientCreator func(grpc.ClientConnInterface) T) *ShardedClientMgr[T] {
	return &ShardedClientMgr[T]{
		ClientCreator: clientCreator,
	}
}

func (ssm *ShardedClientMgr[T]) GetClient(entityId string) (*RpcClient[T], error) {
	return nil, nil
	/*
		var err error = nil
		var shard *Shard
		// find the shard here!
		// Here is where the call to the control service is needed with caching,
		// shard re-assignment checks, load-balancing checks etc
		address := shard.Targets[0].Address
		client, ok := ssm.Addr2Client[address]
		if !ok {
			client, err = NewRpcClient[T](address, ssm.ClientCreator)
			if err == nil {
				ssm.Addr2Client[address] = client
			}
		}
		return client, err
	*/
}
