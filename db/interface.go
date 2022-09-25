package db

import (
	"errors"
)

var UpdateFailed = errors.New("Update failed concurrency check")

type ControlDB interface {
	// Get the target by a given address
	GetTargets(withShards bool, addresses ...string) ([]*Target, error)

	// Updates/Creates a target
	SaveTarget(target *Target) error

	// Deletes a target by address
	DeleteTargets(addresses ...string) error

	// Get all the targets for a given shard.
	GetShard(shardkey string, withTargets bool) (*Shard, error)

	// Create a new shard
	SaveShard(shard *Shard) (err error)

	// For a given shardkey remove all or particular shard assignments
	DeleteShard(shardkey string) (err error)

	// Get all the targets for a given shard key.
	GetShardTargets(shardkey string) ([]*ShardTarget, error)

	// Create a new shard
	SaveShardTarget(shard *ShardTarget) (err error)

	// For a given shardkey remove all or particular shard assignments
	DeleteShardTargets(shardkey string, targets ...string) (err error)
}
