package echosvc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EchoService struct {
	UnimplementedEchoServiceServer

	Prefix    string
	ShardKeys map[string]bool
}

func NewEchoService(prefix string, shardkeys ...string) *EchoService {
	out := &EchoService{
		Prefix:    prefix,
		ShardKeys: make(map[string]bool),
	}
	for _, sk := range shardkeys {
		out.ShardKeys[sk] = true
	}
	return out
}

func (s *EchoService) Echo(ctx context.Context, request *Request) (resp *Response, err error) {
	resp = nil
	err = nil
	if request.Prefix != s.Prefix {
		// The request should not have come here (if sharding did its thing)
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("Can only handle prefix (%s), Provided: %s", s.Prefix, request.Prefix))
	} else {
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("Shard provided: %s", request.Shard))
		for sk, _ := range s.ShardKeys {
			if sk == request.Shard {
				err = nil
				resp = &Response{
					Output: request.Prefix + ":" + request.Shard + ":" + request.Input,
				}
				break
			}
		}
	}
	return
}

func (s *EchoService) UpdateShards(ctx context.Context, request *UpdateShardsRequest) (resp *UpdateShardsResponse, err error) {
	if request.Cmd == "add" {
		for _, sk := range request.Shards {
			s.ShardKeys[sk] = true
		}
	} else {
		for _, sk := range request.Shards {
			s.ShardKeys[sk] = false
		}
	}
	return
}
