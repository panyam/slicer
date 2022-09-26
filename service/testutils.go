package service

import (
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/utils"
	"testing"
)

func CreateTestControlDB(t *testing.T) db.ControlDB {
	// db, dir := OpenSqliteDB(t, forcedir)
	// DB Endpoint eg: postgres://user:pass@localhost:5432/dbname
	dbendpoint := "postgres://postgres:docker@localhost:5432/slicerdb"
	gormdb := utils.OpenPostgresDB(t, dbendpoint)
	gormdb.Where("1 = 1").Delete(&db.Shard{})
	gormdb.Where("1 = 1").Delete(&db.Target{})
	gormdb.Where("1 = 1").Delete(&db.ShardTarget{})
	return db.New(gormdb)
}
