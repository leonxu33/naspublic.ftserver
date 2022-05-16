package server

import (
	"context"
	"fmt"
	"net/http"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/config"
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
		err := server.ListenAndServeTLS(".cert/naspublic.crt", ".cert/naspublic.key")
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
		AllowedHeaders: []string{"Authorization"},
	})
	uploadHandler := uploadCors.Handler(NewUploadHandler())

	// /download
	downloadCors := cors.New(cors.Options{
		AllowedOrigins: config.WebfrontendOrigin,
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		AllowedHeaders: []string{"Authorization"},
	})
	downloadHandler := downloadCors.Handler(NewDownloadHandler())

	// /dir
	dirCors := cors.New(cors.Options{
		AllowedOrigins: config.WebfrontendOrigin,
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowedHeaders: []string{"Authorization"},
	})
	listHandler := dirCors.Handler(NewDirHandler())

	// /auth
	authCors := cors.New(cors.Options{
		AllowedOrigins: config.AuthOrigin,
		AllowedMethods: []string{http.MethodPost, http.MethodGet},
	})
	AuthHandler := authCors.Handler(NewAuthHandler())

	sm := http.NewServeMux()
	sm.Handle(path.Join(config.ApiPath, "upload"), uploadHandler)
	sm.Handle(path.Join(config.ApiPath, "download"), downloadHandler)
	sm.Handle(path.Join(config.ApiPath, "dir"), listHandler)
	sm.Handle(path.Join(config.ApiPath, "auth"), AuthHandler)

	return sm
}

func getServerAddr() string {
	return fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
}
