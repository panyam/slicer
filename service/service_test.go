package service

import (
	"context"
	// "fmt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	fmask "google.golang.org/protobuf/types/known/fieldmaskpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"testing"
	"time"
)

var (
	BG = context.Background
)

func dialer(ctrlSvc *ControlService) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	RegisterControlServiceServer(server, ctrlSvc)
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func RunTest(t *testing.T, testBody func(ctx context.Context, client ControlServiceClient)) {
	db := CreateTestControlDB(t)
	ctrlSvc := NewControlService(db)
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(ctrlSvc)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewControlServiceClient(conn)
	testBody(ctx, client)
}

func TestGetTarget_InvalidTarget(t *testing.T) {
	RunTest(t, func(ctx context.Context, client ControlServiceClient) {
		target, err := client.GetTarget(BG(), &GetTargetRequest{
			Address: "test",
		})
		if er, ok := status.FromError(err); ok {
			assert.Equal(t, er.Code(), codes.NotFound)
		}
		assert.Equal(t, target, (*Target)(nil))
	})
}

func TestCreateTarget_InvalidTarget(t *testing.T) {
	RunTest(t, func(ctx context.Context, client ControlServiceClient) {
		target, err := client.GetTarget(BG(), &GetTargetRequest{
			Address: "test",
		})
		if er, ok := status.FromError(err); ok {
			assert.Equal(t, er.Code(), codes.NotFound)
		}
		assert.Equal(t, target, (*Target)(nil))

		// now create
		target, err = client.UpdateTarget(BG(), &UpdateTargetRequest{
			Target: &Target{
				Address:  "test",
				Status:   "ACTIVE",
				PingedAt: tspb.New(time.Now()),
			},
			UpdateMask: &fmask.FieldMask{},
		})
		assert.Equal(t, err, nil)
		assert.Equal(t, target, (*Target)(nil))
	})
}
