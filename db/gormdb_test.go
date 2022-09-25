package db

import (
	"fmt"
	"github.com/panyam/slicer/utils"
	"github.com/stretchr/testify/assert"
	"log"
	// "reflect"
	"time"
	// "log"
	// "os"
	"testing"
)

func CreateTestDB(t *testing.T) ControlDB {
	// db, dir := OpenSqliteDB(t, forcedir)
	// DB Endpoint eg: postgres://user:pass@localhost:5432/dbname
	dbendpoint := "postgres://postgres:docker@localhost:5432/slicerdb"
	db := utils.OpenPostgresDB(t, dbendpoint)
	db.Where("1 = 1").Delete(&Shard{})
	db.Where("1 = 1").Delete(&Target{})
	db.Where("1 = 1").Delete(&ShardTarget{})
	return New(db)
}

func TestNewDB(t *testing.T) {
	CreateTestDB(t)
}

func TestSaveNewTarget(t *testing.T) {
	db := CreateTestDB(t)

	g1, err := db.GetTargets(false, "addr1")
	assert.Equal(t, len(g1), 0)
	assert.Equal(t, err.Error(), "record not found")

	gs, err := db.GetTargets(false, "addr1", "addr2")
	assert.Equal(t, len(gs), 0)
	assert.Equal(t, err, nil)

	// Create a couple and see what happens
	t1 := &Target{
		Address: "addr1",
		Status:  "ACTIVE",
		Version: 1,
	}
	err = db.SaveTarget(t1)
	assert.Equal(t, err, nil)
	t2, err := db.GetTargets(false, "addr1")
	t1.CreatedAt = t2[0].CreatedAt
	t1.UpdatedAt = t2[0].UpdatedAt
	t1.PingedAt = t2[0].PingedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2[0], t1)
	assert.Equal(t, t2[0].Version, 2)

	// now update it normally
	t1.Status = "PENDING"
	t1.PingedAt = time.Now()
	err = db.SaveTarget(t1)
	assert.Equal(t, err, nil)
	t2, err = db.GetTargets(false, "addr1")
	t1.UpdatedAt = t2[0].UpdatedAt
	assert.Equal(t, t2[0].PingedAt.Equal(t1.PingedAt), true)
	t1.PingedAt = t2[0].PingedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2[0].Version, 3)
	assert.Equal(t, t2[0], t1)

	// Try to ensure no false updates
	t3 := *t1
	t3.Status = "ACTIVE"
	t3.Version = t3.Version + 100
	err = db.SaveTarget(&t3)
	assert.Equal(t, err, UpdateFailed)
	t2, err = db.GetTargets(false, "addr1")
	// t1.UpdatedAt = t2[0].UpdatedAt
	// assert.Equal(t, t2[0].PingedAt.Equal(t1.PingedAt), true)
	// t1.PingedAt = t2[0].PingedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2[0].Version, 3) // no update should happen
	assert.Equal(t, t2[0], t1)
}

func TestSaveNewShard(t *testing.T) {
	db := CreateTestDB(t)

	g1, err := db.GetShard("key1", false)
	log.Println("G1, err: ", g1, err)
	assert.Equal(t, g1, (*Shard)(nil))
	assert.Equal(t, err, nil)

	// Create a couple and see what happens
	t1 := &Shard{
		Key:     "key1",
		Version: 1,
	}
	err = db.SaveShard(t1)
	assert.Equal(t, err, nil)
	t2, err := db.GetShard("key1", false)
	t1.CreatedAt = t2.CreatedAt
	t1.UpdatedAt = t2.UpdatedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2, t1)
	assert.Equal(t, t2.Version, 2)

	// now update it normally
	err = db.SaveShard(t1)
	assert.Equal(t, err, nil)
	t2, err = db.GetShard("key1", false)
	t1.UpdatedAt = t2.UpdatedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2.Version, 3)
	assert.Equal(t, t2, t1)

	// Try to ensure no false updates
	t3 := *t1
	t3.Version = t3.Version + 100
	err = db.SaveShard(&t3)
	assert.Equal(t, err, UpdateFailed)
	t2, err = db.GetShard("key1", false)
	// t1.UpdatedAt = t2.UpdatedAt
	// assert.Equal(t, t2.PingedAt.Equal(t1.PingedAt), true)
	// t1.PingedAt = t2.PingedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2.Version, 3) // no update should happen
	assert.Equal(t, t2, t1)
}

func TestSaveNewShardTarget(t *testing.T) {
	db := CreateTestDB(t)

	// Create 10 shards
	for i := 0; i < 10; i++ {
		db.SaveShard(&Shard{
			Key: fmt.Sprintf("key%d", i),
		})
	}

	// create a few targets
	for i := 0; i < 10; i++ {
		db.SaveTarget(&Target{
			Address: fmt.Sprintf("addr%d", i),
			Status:  "ACTIVE",
		})
	}

	db.SaveShardTarget(&ShardTarget{
		ShardKey:      "key1",
		TargetAddress: "addr1",
		Status:        "ACTIVE",
	})

	db.SaveShardTarget(&ShardTarget{
		ShardKey:      "key1",
		TargetAddress: "addr2",
		Status:        "ACTIVE",
	})

	db.SaveShardTarget(&ShardTarget{
		ShardKey:      "key1",
		TargetAddress: "addr3",
		Status:        "ACTIVE",
	})

	db.SaveShardTarget(&ShardTarget{
		ShardKey:      "key2",
		TargetAddress: "addr3",
		Status:        "ACTIVE",
	})

	db.SaveShardTarget(&ShardTarget{
		ShardKey:      "key2",
		TargetAddress: "addr4",
		Status:        "ACTIVE",
	})

	db.SaveShardTarget(&ShardTarget{
		ShardKey:      "key2",
		TargetAddress: "addr5",
		Status:        "ACTIVE",
	})

	targets, err := db.GetShardTargets("key1")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(targets), 3)
	assert.Equal(t, targets[0].TargetAddress, "addr1")
	assert.Equal(t, targets[1].TargetAddress, "addr2")
	assert.Equal(t, targets[2].TargetAddress, "addr3")
}
