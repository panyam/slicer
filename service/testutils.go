package control

import (
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/utils"
	"testing"
)

func CreateTestControlDB(t *testing.T) db.ControlDB {
	// db, dir := OpenSqliteDB(t, forcedir)
	// DB Endpoint eg: postgres://user:pass@localhost:5432/dbname
	dbendpoint := "postgres://postgres:docker@localhost:5432/slicerdb"
	db := utils.OpenPostgresDB(t, dbendpoint)
	db.Where("1 = 1").Delete(&Shard{})
	db.Where("1 = 1").Delete(&Target{})
	return NewControlDB(db)
}
