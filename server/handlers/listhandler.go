package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/config"
	"github.com/lyokalita/naspublic.ftserver/server/filebrowser"
)

type ListHandler struct {
}

func NewListHandler() *ListHandler {
	return &ListHandler{}
}

func (hdl *ListHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		hdl.handleGet(rw, r)
		return
	}
	rw.WriteHeader(http.StatusMethodNotAllowed)
}

func (hdl *ListHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	log.Debug("handle list file request")

	keys, ok := r.URL.Query()["key"]
	queryDir := ""
	if ok && len(keys[0]) > 0 {
		queryDir = keys[0]
	}

	_, err := os.Stat(path.Join(config.PublicDirectoryRoot, queryDir))
	if err != nil {
		log.Infof("Directory %s does not exit", queryDir)
		http.Error(rw, "Directory does not exit", 404)
		return
	}

	res := &ListFileResponse{}
	metadataList, err := filebrowser.GetFileMetadataList(queryDir)
	if err != nil {
		res.Status = -1
		res.Message = "failed to list files"
		log.Errorf("failed to list download files, err: %v", err)
	}
	res.Message = "success"
	res.Data = &ListFileResponseData{
		QueryFolder:  queryDir,
		MetadataList: metadataList,
	}
	res.ToJSON(rw)
	log.Infof("response value - success: %d, message: %s, count: %d", res.Status, res.Message, len(res.Data.MetadataList))
}

type ListFileResponse struct {
	Status  int                   `json:"status"`
	Message string                `json:"message"`
	Data    *ListFileResponseData `json:"data"`
}

type ListFileResponseData struct {
	QueryFolder  string                      `json:"queryFolder"`
	MetadataList []*filebrowser.FileMetadata `json:"metadatas"`
}

func (p *ListFileResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}
