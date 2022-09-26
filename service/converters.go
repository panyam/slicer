package service

import (
	// gut "github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/protos"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func ShardTargetToProto(input *db.ShardTarget) (out *protos.ShardTarget) {
	out = &protos.ShardTarget{
		Status:    input.Status,
		UpdatedAt: tspb.New(input.UpdatedAt),
		CreatedAt: tspb.New(input.CreatedAt),
		Target:    input.TargetAddress,
	}
	return
}

func ShardTargetFromProto(input *protos.ShardTarget) (out *db.ShardTarget) {
	out = &db.ShardTarget{
		TargetAddress: input.Target,
		Status:        input.Status,
		UpdatedAt:     input.UpdatedAt.AsTime(),
		CreatedAt:     input.CreatedAt.AsTime(),
	}
	return
}

func TargetToProto(input *db.Target) (out *protos.Target) {
	out = &protos.Target{
		Address:   input.Address,
		Status:    input.Status,
		UpdatedAt: tspb.New(input.UpdatedAt),
	}
	return
}

func TargetFromProto(input *protos.Target) (out *db.Target) {
	out = &db.Target{
		Address:   input.Address,
		Status:    input.Status,
		UpdatedAt: input.UpdatedAt.AsTime(),
	}
	return
}
