package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/config"
)

type DownloadHandler struct {
}

func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{}
}

func (hdl *DownloadHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		hdl.handleGet(rw, r)
		return
	}
	rw.WriteHeader(http.StatusMethodNotAllowed)
}

func (hdl *DownloadHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	log.Debug("handle file download request")
	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		log.Errorf("Url Param 'key'is missing")
		return
	}
	key := keys[0]

	sourcePath := path.Join(config.PublicDirectoryRoot, key)
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", key))
	http.ServeFile(rw, r, sourcePath)
	log.Infof("file %s served", sourcePath)
}

func getContentType(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return "application/octet-stream"
	}
	defer f.Close()
	buffer := make([]byte, 512)
	f.Read(buffer)
	return http.DetectContentType(buffer)
}
