package harness

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panyam/slicer/clients"
	"github.com/panyam/slicer/cmd/echosvc"
	"github.com/panyam/slicer/protos"
	"github.com/panyam/slicer/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
	"log"
	"net/http"
	"os"
	// "os/signal"
	"sync"
	"syscall"
	"time"
)

/**
 * A simple binary that acts as a shard provider.
 * Shard providers are services that are running performing a simple function on its input.
 * They append the provider prefix to the input (which is the shard key).
 * The providers are only allowed to serve requests
 * if they own a particular shard.
 */

type WebClient struct {
	Addr        string
	ControlAddr string
	Router      *gin.Engine
	quitChan    chan os.Signal
	wg          sync.WaitGroup
	Logger      *log.Logger
}

func NewWebClient(addr string, logfile io.Writer) *WebClient {
	out := WebClient{
		Addr:        addr,
		ControlAddr: "localhost:7000",
		Router:      gin.Default(),
		Logger:      log.New(logfile, fmt.Sprintf("CLNT:[%s]", addr), log.Ldate|log.Ltime|log.Lshortfile),
	}
	out.Router.Use(gin.LoggerWithWriter(logfile))
	return &out
}

func (w *WebClient) Stop() {
	w.quitChan <- syscall.SIGTERM
}

func (w *WebClient) Start() {
	// Updating the controller itself
	w.Router.GET("/control/:prefix/:shard/:input/", func(ctx *gin.Context) {
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
	clientMgr := clients.NewStaticClientMgr(w.ControlAddr, protos.NewControlServiceClient)
	ctrlSvcClient, err := clientMgr.GetClient("")
	w.Logger.Println("CC, E: ", ctrlSvcClient, err)
	if err != nil {
		panic(err)
	}
	clientMap := make(map[string]*clients.RpcClient[echosvc.EchoServiceClient])

	w.Router.GET("/:prefix/:shard/:input/", func(ctx *gin.Context) {
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
				w.sendResponse(ctx, nil, err)
				return
			}

			// our resp has a bunch of addresses - see if we have those here
			// create a client against this shard's address
			address := resp.Targets[0].Target
			newClient, err := clients.NewRpcClient(address, echosvc.NewEchoServiceClient)
			if err != nil {
				w.sendResponse(ctx, nil, err)
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
		w.sendResponse(ctx, response, err)
	})

	srv := &http.Server{
		Addr:    w.Addr,
		Handler: w.Router,
	}

	w.wg.Add(2)
	go func() {
		defer w.wg.Done()
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			w.Logger.Printf("WebClient (%s) Listen: %s\n", w.Addr, err)
		}
	}()

	// Listener for the cancel
	w.quitChan = make(chan os.Signal)
	go func() {
		defer w.wg.Done()
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		// kill (no param) default send syscall.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
		// signal.Notify(w.quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-w.quitChan
		w.Logger.Printf("Shutting down web client: %s", w.Addr)

		// The context is used to inform the server it has 5 seconds to finish
		// the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			w.Logger.Println(fmt.Sprintf("Server (%s) forced to shutdown: ", w.Addr), err)
		}
	}()
}

func (w *WebClient) sendResponse(ctx *gin.Context, resp protoreflect.ProtoMessage, err error) {
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
