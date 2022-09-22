package db

import (
	"errors"
)

var UpdateFailed = errors.New("Update failed concurrency check")

type ControlDB interface {
	// Get the target by a given address
	GetTargets(addresses ...[]string) ([]*Target, error)

	// Updates/Creates a target
	SaveTarget(target *Target) error

	// Deletes a target by address
	DeleteTargets(addresses ...[]string) error

	// Get all the targets for a given shard.
	GetShards(shardkey string) (*[]Shard, error)

	// Create a new shard
	SaveShard(shard *Shard) (err error)

	// For a given shardkey remove all or particular shard assignments
	DeleteShard(shardkey string, targets ...[]string) (err error)
}
