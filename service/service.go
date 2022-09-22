package control

import (
	"context"
	"fmt"
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

func (s *ControlService) GetTarget(ctx context.Context, request *GetTargetRequest) (resp *Target, err error) {
	resp = nil
	err = nil
	target, err := s.ControlDB.GetTarget(request.Address)
	if target == nil {
		resp = nil
		err = status.Error(codes.NotFound, fmt.Sprintf("Target not found: %s", request.Address))
	} else {
		resp = TargetToProto(target)
	}
	return
}

func (s *ControlService) GetShard(ctx context.Context, request *GetShardRequest) (resp *GetShardResponse, err error) {
	resp = nil
	err = nil
	shard, err := s.ControlDB.GetShard(request.Shard.Key)
	if shard == nil {
		resp = &GetShardResponse{}
	} else {
		resp = &GetShardResponse{
			Shard: ShardToProto(shard),
		}
	}
	return
}

func (s *ControlService) UpdateShard(ctx context.Context, request *UpdateShardRequest) (resp *UpdateShardResponse, err error) {
	return
}
