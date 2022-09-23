package echosvc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type EchoService struct {
	UnimplementedEchoServiceServer

	Prefix    string
	ShardKeys []string
}

func NewEchoService(prefix string, shardkeys ...string) *EchoService {
	return &EchoService{
		Prefix:    prefix,
		ShardKeys: shardkeys,
	}
}

func (s *EchoService) Echo(ctx context.Context, request *Request) (resp *Response, err error) {
	resp = nil
	err = nil
	if request.Prefix != s.Prefix {
		// The request should not have come here (if sharding did its thing)
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("Can only handle prefix (%s), Provided: %s", s.Prefix, request.Prefix))
	} else {
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("Shard provided: %s, Can only handle shards: %s", request.Shard, strings.Join(s.ShardKeys, ", ")))
		for _, sk := range s.ShardKeys {
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
