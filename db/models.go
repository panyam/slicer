package db

import (
	"github.com/lib/pq"
	"time"
)

type Target struct {
	Address   string `gorm:"primaryKey"`
	Status    string
	Tags      pq.StringArray `gorm:"type:text[]" gorm:"index:ByTag"`
	UpdatedAt time.Time
	PingedAt  time.Time
	Version   int // used for optimistic locking
}

type Shard struct {
	ShardKey      string `gorm:"primaryKey" gorm:"index:ByShard"`
	TargetAddress string `gorm:"primaryKey" gorm:"index:ByTarget"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Status        string
	Version       int // used for optimistic locking
}
