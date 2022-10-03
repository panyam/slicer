package service

import (
	"context"
	"fmt"
	gut "github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"sort"
	"time"
)

type ControlService struct {
	protos.UnimplementedControlServiceServer
	ControlDB db.ControlDB
	Logger    *log.Logger
}

func NewControlService(ctrldb db.ControlDB, logger *log.Logger) (out *ControlService) {
	out = &ControlService{
		ControlDB: ctrldb,
		Logger:    logger,
	}
	return
}

func (s *ControlService) GetTargets(ctx context.Context, request *protos.GetTargetsRequest) (resp *protos.GetTargetsResponse, err error) {
	resp = nil
	err = nil
	targets, err := s.ControlDB.GetTargets(true, request.Address...)
	if targets == nil {
		resp = nil
		err = status.Error(codes.NotFound, fmt.Sprintf("Target not found: %s", request.Address))
	} else {
		resp = &protos.GetTargetsResponse{
			Targets: gut.Map(targets, TargetToProto),
		}
	}
	return
}

func (s *ControlService) PingTarget(ctx context.Context, request *protos.PingTargetRequest) (resp *protos.Target, err error) {
	targets, err := s.ControlDB.GetTargets(false, request.Address)
	if err == nil {
		if len(targets) == 0 {
			// create it
			newtarget := &db.Target{
				Address:  request.Address,
				PingedAt: time.Now(),
				Status:   "ACTIVE",
			}
			err = s.ControlDB.SaveTarget(newtarget)
			s.Logger.Println("Trying to save: ", newtarget, err)
			if err == nil {
				resp = TargetToProto(newtarget)
			}
		} else {
			targets[0].PingedAt = time.Now()
			err = s.ControlDB.SaveTarget(targets[0])
			if err == nil {
				resp = TargetToProto(targets[0])
			} else {
				s.Logger.Println("SaveTarget error: ", err)
			}
		}
	}
	return resp, err
}

func (s *ControlService) SaveTarget(ctx context.Context, request *protos.SaveTargetRequest) (resp *protos.Target, err error) {
	dbt := TargetFromProto(request.Target)
	err = s.ControlDB.SaveTarget(dbt)
	if err != nil {
		resp = TargetToProto(dbt)
	}
	return resp, err
}

func (s *ControlService) DeleteTargets(ctx context.Context, request *protos.DeleteTargetsRequest) (resp *protos.DeleteTargetsResponse, err error) {
	err = s.ControlDB.DeleteTargets(request.Addresses...)
	return nil, err
}

// Shard specific APIs

func (s *ControlService) GetShard(ctx context.Context, request *protos.GetShardRequest) (resp *protos.GetShardResponse, err error) {
	resp = nil
	err = nil
	shard_targets, err := s.ControlDB.GetShardTargets(request.Shard.Key)
	if err == nil {
		addresses := gut.Map(shard_targets, func(st *db.ShardTarget) string {
			return st.TargetAddress
		})
		var targets []*db.Target
		targets, err = s.ControlDB.GetTargets(false, addresses...)
		if err == nil {
			target_info := make(map[string]*protos.Target)
			for _, target := range targets {
				target_info[target.Address] = TargetToProto(target)
			}
			// Order Shard targets by most recently pinged
			sort.Slice(shard_targets, func(t1 int, t2 int) bool {
				return targets[t2].PingedAt.Sub(targets[t1].PingedAt) < 0
			})
			resp = &protos.GetShardResponse{
				Key:        &protos.ShardKey{Key: request.Shard.Key},
				Targets:    gut.Map(shard_targets, ShardTargetToProto),
				TargetInfo: target_info,
			}
		}
	}
	return
}

func (s *ControlService) SaveShard(ctx context.Context, request *protos.SaveShardRequest) (resp *protos.SaveShardResponse, err error) {
	// TODO - Transactionalize this
	if request.RemoveAll {
		err = s.ControlDB.DeleteShardTargets(request.Key.Key)
	} else {
		err = s.ControlDB.DeleteShardTargets(request.Key.Key, request.Removed...)
	}
	if err == nil {
		for _, address := range request.Added {
			err = s.ControlDB.SaveShardTarget(&db.ShardTarget{
				ShardKey:      request.Key.Key,
				TargetAddress: address,
				Status:        "ACTIVE",
			})
		}
	}
	return
}

func (s *ControlService) DeleteShard(ctx context.Context, request *protos.DeleteShardRequest) (resp *protos.DeleteShardResponse, err error) {
	err = s.ControlDB.DeleteShardTargets(request.Key.Key, request.Addresses...)
	return nil, err
}
