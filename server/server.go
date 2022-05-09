package server

import (
	"context"
	"fmt"
	"net/http"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/config"
	"github.com/lyokalita/naspublic.ftserver/server/handlers"
	"github.com/rs/cors"
)

var server *http.Server

func StartHttpServer() {
	sm := constructServerMux()
	addr := getServerAddr()

	server = &http.Server{
		Addr:    addr,
		Handler: sm,
		// IdleTimeout:  time.Duration(120) * time.Second,
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 5 * time.Second,
	}

	// wrapping ListenAndServe in gofunc so it's not going to block
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Error(err)
		}
	}()

	log.Infof("nas file transfer server listens at %s", path.Join(addr, config.ApiPath))
}

func StopHttpServer(ctx context.Context) {
	err := server.Shutdown(ctx)
	if err != nil {
		log.Info(err)
	}
}

func constructServerMux() *http.ServeMux {
	// /upload
	uploadCors := cors.New(cors.Options{
		AllowedOrigins: config.WebfrontendOrigin,
		AllowedMethods: []string{http.MethodPost},
	})
	uploadHandler := uploadCors.Handler(handlers.NewUploadHandler())

	// /list
	listCors := cors.New(cors.Options{
		AllowedOrigins: config.WebfrontendOrigin,
		AllowedMethods: []string{http.MethodGet},
	})
	listHandler := listCors.Handler(handlers.NewListHandler())

	// /download
	downloadCors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet},
	})
	downloadHandler := downloadCors.Handler(handlers.NewDownloadHandler())
	sm := http.NewServeMux()
	sm.Handle(path.Join(config.ApiPath, "upload"), uploadHandler)
	sm.Handle(path.Join(config.ApiPath, "list"), listHandler)
	sm.Handle(path.Join(config.ApiPath, "download"), downloadHandler)
	return sm
}

func getServerAddr() string {
	return fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
}
