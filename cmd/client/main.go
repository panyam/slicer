package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/panyam/slicer/clients"
	"github.com/panyam/slicer/cmd/echosvc"
	"github.com/panyam/slicer/protos"
	"github.com/panyam/slicer/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"log"
	"net/http"
)

/**
 * A simple binary that acts as a shard provider.
 * Shard providers are services that are running performing a simple function on its input.
 * They append the provider prefix to the input (which is the shard key).
 * The providers are only allowed to serve requests
 * if they own a particular shard.
 */

var (
	addr         = flag.String("addr", "localhost:8000", "Address to run the echo web proxy client on.")
	control_addr = flag.String("control_addr", "localhost:7000", "Address where control service is running.")
)

func sendResponse(ctx *gin.Context, resp protoreflect.ProtoMessage, err error) {
	if err != nil {
		if er, ok := status.FromError(err); ok {
			code := er.Code()
			msg := er.Message()
			httpCode := http.StatusInternalServerError
			// see if we have a specific client error
			if code == codes.PermissionDenied {
				httpCode = http.StatusForbidden
			} else if code == codes.NotFound {
				httpCode = http.StatusNotFound
			} else if code == codes.AlreadyExists {
				httpCode = http.StatusConflict
			} else if code == codes.InvalidArgument {
				httpCode = http.StatusBadRequest
			}
			ctx.JSON(httpCode, gin.H{"error": code, "message": msg})
		}
	} else {
		jsonData := utils.ProtoToJson(resp)
		ctx.Data(http.StatusOK, gin.MIMEJSON, jsonData)
	}
}

func main() {
	flag.Parse()
	router := gin.Default()

	// observer := NewObserver()
	// ctrl := observer.DiscoverContoller()

	// Updating the controller itself
	router.GET("/control/:prefix/:shard/:input/", func(ctx *gin.Context) {
		/*
			shard := ctx.Param("shard")
			client := clientmgr.GetClient(shard)
			response, err := client.Echo(&Request{
				Prefix: ctx.Param("prefix"),
				Shard:  shard,
				Input:  ctx.Param("input"),
			})
			sendResponse(response)
		*/
	})

	// This would be replaced by the discovery service otherwise so that
	// every node would have the "latest" client
	clientMgr := clients.NewStaticClientMgr(*control_addr, protos.NewControlServiceClient)
	ctrlSvcClient, err := clientMgr.GetClient("")
	log.Println("CC, E: ", ctrlSvcClient, err)
	if err != nil {
		panic(err)
	}
	clientMap := make(map[string]*clients.RpcClient[echosvc.EchoServiceClient])

	router.GET("/:prefix/:shard/:input/", func(ctx *gin.Context) {
		shard := ctx.Param("shard")
		client, ok := clientMap[shard]
		if !ok {
			resp, err := ctrlSvcClient.Client.GetShard(
				context.Background(),
				&protos.GetShardRequest{
					Shard: &protos.ShardKey{Key: shard},
				},
			)
			if err != nil {
				sendResponse(ctx, nil, err)
				return
			}

			// our resp has a bunch of addresses - see if we have those here
			// create a client against this shard's address
			address := resp.Targets[0].Target
			newClient, err := clients.NewRpcClient(address, echosvc.NewEchoServiceClient)
			if err != nil {
				sendResponse(ctx, nil, err)
				return
			}
			clientMap[shard] = newClient
			client = newClient
		}

		response, err := client.Client.Echo(context.Background(), &echosvc.Request{
			Prefix: ctx.Param("prefix"),
			Shard:  shard,
			Input:  ctx.Param("input"),
		})
		sendResponse(ctx, response, err)
	})
	router.Run(*addr)
}
