package harness

import (
	"fmt"
	"github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/protos"
	"github.com/panyam/slicer/service"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net"
	"strings"
)

type Controller struct {
	Addr       string
	DBEndpoint string
	grpcServer *grpc.Server
	ctrldb     db.ControlDB
	Logger     *log.Logger
}

func NewController(addr string, db_endpoint string, logfile io.Writer) *Controller {
	out := Controller{
		Addr:       addr,
		DBEndpoint: db_endpoint,
		grpcServer: grpc.NewServer(
		// grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		// grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		),
		Logger: log.New(logfile, fmt.Sprintf("CTRL:[%s]", addr), log.Ldate|log.Ltime|log.Lshortfile),
	}
	gormdb, err := OpenDB(db_endpoint)
	if err != nil {
		panic(err)
	}
	out.ctrldb = db.New(gormdb)
	// grpc_prometheus.Register(grpcServer)
	return &out
}

func (c *Controller) Stop() {
	c.grpcServer.Stop()
}

func (c *Controller) Start() {
	protos.RegisterControlServiceServer(c.grpcServer, service.NewControlService(c.ctrldb))
	c.Logger.Printf("Initializing Control Server on %s", c.Addr)
	lis, err := net.Listen("tcp", c.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	c.grpcServer.Serve(lis)
}

func OpenDB(db_endpoint string) (db *gorm.DB, err error) {
	var dbpath string
	if strings.HasPrefix(db_endpoint, "sqlite://") {
		dbpath = utils.ExpandUserPath((db_endpoint)[len("sqlite://"):])
		db, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	} else if strings.HasPrefix(db_endpoint, "postgres://") {
		db, err = gorm.Open(postgres.Open(db_endpoint), &gorm.Config{})
	}
	if err != nil {
		log.Printf("Cannot connect DB: %s", db_endpoint)
	}
	return
}
