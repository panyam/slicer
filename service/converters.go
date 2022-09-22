package control

import (
	gut "github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func ShardToProto(input *db.Shard) (out *Shard) {
	targets := gut.Map(input.Targets, TargetToProto)
	out = &Shard{
		UpdatedAt: tspb.New(input.UpdatedAt),
		Key:       &ShardKey{Key: input.Key},
		Targets:   targets,
	}
	return
}

func ShardFromProto(input *Shard) (out *db.Shard) {
	targets := gut.Map(input.Targets, TargetFromProto)
	out = &Shard{
		Key:       input.Key.Key,
		UpdatedAt: input.UpdatedAt.AsTime(),
		Targets:   targets,
	}
	return
}

func TargetToProto(input *db.Target) (out *Target) {
	out = &Target{
		Address:   input.Address,
		Status:    input.Status,
		UpdatedAt: tspb.New(input.UpdatedAt),
	}
	return
}

func TargetFromProto(input *Target) (out *db.Target) {
	out = &Target{
		Address:   input.Address,
		Status:    input.Status,
		UpdatedAt: input.UpdatedAt.AsTime(),
	}
	return
}
