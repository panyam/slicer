package db

import (
	// "fmt"
	"github.com/panyam/slicer/utils"
	"github.com/stretchr/testify/assert"
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

	g1, err := db.GetTargets("addr1")
	assert.Equal(t, len(g1), 0)
	assert.Equal(t, err.Error(), "record not found")

	gs, err := db.GetTargets("addr1", "addr2")
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
	t2, err := db.GetTargets("addr1")
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
	t2, err = db.GetTargets("addr1")
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
	t2, err = db.GetTargets("addr1")
	// t1.UpdatedAt = t2[0].UpdatedAt
	// assert.Equal(t, t2[0].PingedAt.Equal(t1.PingedAt), true)
	// t1.PingedAt = t2[0].PingedAt
	assert.Equal(t, err, nil)
	assert.Equal(t, t2[0].Version, 3) // no update should happen
	assert.Equal(t, t2[0], t1)
}
