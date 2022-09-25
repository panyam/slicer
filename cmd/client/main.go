package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/panyam/slicer/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	addr = flag.String("addr", "localhost:8000", "Address to run the echo web proxy client on.")
)

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

	router.GET("/:prefix/:shard/:input/", func(ctx *gin.Context) {
		/*
			shard := ctx.Param("shard")
			client := clientmgr.Get(shard)
			response, err := client.Echo(&Request{
				Prefix: ctx.Param("prefix"),
				Shard:  shard,
				Input:  ctx.Param("input"),
			})
			sendResponse(response)
		*/
	})
	router.Run(*addr)
}

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
