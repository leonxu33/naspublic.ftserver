package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/config"
	"github.com/lyokalita/naspublic.ftserver/server"
)

func main() {
	// load configuration
	rand.Seed(time.Now().UnixNano())
	config.Init()
	defer log.Flush()
	log.Info("successfully initialized application")

	// create http server
	server.StartHttpServer()

	// make a new channel to notify on os interrupt of server (ctrl + C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)

	// This blocks the code until the channel receives some message
	sig := <-sigChan
	log.Info("received terminate, graceful shutdown", sig)
	// Once message is consumed shut everything down
	// Gracefully shuts down all client requests. Makes server more reliable
	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.StopHttpServer(tc)
}
