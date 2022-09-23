package main

import (
	"flag"
	"github.com/gin-gonic/gin"
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
	router := gin.Default()

	router.GET("/:prefix/:shard/:input/", func(ctx *gin.Context) {
		// TODO -- Find the right client and send it and return the response
	})
	router.Run(*addr)
}
