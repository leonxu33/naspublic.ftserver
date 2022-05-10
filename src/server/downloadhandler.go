package server

import (
	"fmt"
	"net/http"
	"os"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
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
	key := path.Join(keys[0])

	sourcePath := path.Join(config.PublicDirectoryRoot, key)
	if !utils.IsPathValid(sourcePath) {
		log.Warnf("invalid query, %s", sourcePath)
		http.Error(rw, "invalid query", 404)
		return
	}

	info, err := os.Stat(sourcePath)
	if err != nil || info.IsDir() {
		log.Warnf("file does not exist, %s", sourcePath)
		http.Error(rw, "File does not exist", 404)
		return
	}

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
