package utils

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
)

func OpenSqliteDB(t *testing.T, forcedir string) (*gorm.DB, string) {
	var err error
	var dir string
	if forcedir != "" {
		dir = forcedir
		if err = os.RemoveAll(forcedir); err != nil {
			log.Println("Cannot Remove dir: ", err)
		}
		if err = os.MkdirAll(forcedir, os.ModePerm); err != nil {
			log.Println("Cannot Create dir: ", err)
		}
	} else {
		dir, err = ioutil.TempDir("/tmp", "slicerdb")
		if err != nil {
			log.Fatal(err)
		}
	}

	filepath := path.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	fmt.Println("DBRoot: ", filepath)
	return db, dir
}

func OpenPostgresDB(t *testing.T, dbendpoint string) *gorm.DB {
	// DB Endpoint eg: postgres://user:pass@localhost:5432/dbname
	db, err := gorm.Open(postgres.Open(dbendpoint), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
