package main

import (
	"flag"
	"github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/protos"
	"github.com/panyam/slicer/service"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net"
	"strings"
)

var (
	addr        = flag.String("addr", "localhost:7000", "Address to run the echo service on.")
	db_endpoint = flag.String("db_endpoint", "postgres://postgres:docker@localhost:5432/slicerdb", "Endpoint of DB backing slicer shard targets.  Supported - sqlite eg (sqlite://~/.slicer/sqlite.db) or postgres eg (postgres://user:pass@localhost:5432/dbname)")
)

func main() {
	flag.Parse()
	grpcServer := grpc.NewServer(
	// grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	// grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	gormdb, err := OpenDB(*db_endpoint)
	if err != nil {
		panic(err)
	}
	ctrldb := db.New(gormdb)
	protos.RegisterControlServiceServer(grpcServer, service.NewControlService(ctrldb))
	// grpc_prometheus.Register(grpcServer)
	log.Printf("Initializing Control Server on %s", *addr)
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer.Serve(lis)
}

func OpenDB(db_endpoint string) (db *gorm.DB, err error) {
	var dbpath string
	if strings.HasPrefix(db_endpoint, "sqlite://") {
		dbpath = utils.ExpandUserPath((db_endpoint)[len("sqlite://"):])
		db, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	} else if strings.HasPrefix(db_endpoint, "postgres://") {
		db, err = gorm.Open(postgres.Open(db_endpoint), &gorm.Config{})
	}
	log.Printf("Cannot connect DB: %s", db_endpoint)
	return
}
