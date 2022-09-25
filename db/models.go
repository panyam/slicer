package db

import (
	"github.com/lib/pq"
	"time"
)

type Shard struct {
	Key       string    `gorm:"primaryKey"`
	Targets   []*Target `gorm:"many2many:shard_targets;"`
	Version   int       // used for optimistic locking
	CreatedAt time.Time
	UpdatedAt time.Time
	PingedAt  time.Time
}

type Target struct {
	Address   string `gorm:"primaryKey"`
	Status    string
	Tags      pq.StringArray `gorm:"type:text[]" gorm:"index:ByTag"`
	Shards    []*Shard       `gorm:"many2many:shard_targets;"`
	Version   int            // used for optimistic locking
	CreatedAt time.Time
	UpdatedAt time.Time
	PingedAt  time.Time
}

type ShardTarget struct {
	ShardKey      string `gorm:"primaryKey" gorm:"index:ByShard"`
	TargetAddress string `gorm:"primaryKey" gorm:"index:ByTarget"`
	Status        string `gorm:"index"`
	Version       int    // used for optimistic locking
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
