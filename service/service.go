package control

import (
	"context"
	"fmt"
	gut "github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ControlService struct {
	UnimplementedControlServiceServer
	ControlDB db.ControlDB
}

func NewControlService(ctrldb db.ControlDB) (out *ControlService) {
	out = &ControlService{
		ControlDB: ctrldb,
	}
	return
}

func (s *ControlService) GetTargets(ctx context.Context, request *GetTargetsRequest) (resp *GetTargetsResponse, err error) {
	resp = nil
	err = nil
	targets, err := s.ControlDB.GetTargets(request.Address)
	if targets == nil {
		resp = nil
		err = status.Error(codes.NotFound, fmt.Sprintf("Target not found: %s", request.Address))
	} else {
		resp = &GetTargetsResponse{
			Targets: gut.Map(targets, TargetToProto),
		}
	}
	return
}

func (s *ControlService) GetShard(ctx context.Context, request *GetShardRequest) (resp *GetShardResponse, err error) {
	resp = nil
	err = nil
	shards, err := s.ControlDB.GetShards(request.Shard.Key)
	if shards == nil {
		resp = &GetShardResponse{}
	} else {
		resp = &GetShardResponse{
			Shard: gut.Map(shards, ShardToProto),
		}
	}
	return
}

func (s *ControlService) SaveShard(ctx context.Context, request *SaveShardRequest) (resp *SaveShardResponse, err error) {
	return
}
