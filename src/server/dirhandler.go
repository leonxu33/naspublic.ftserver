package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/fs"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
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
	if r.Method == http.MethodPost {
		hdl.handlePost(rw, r)
		return
	}
	if r.Method == http.MethodDelete {
		hdl.handleDelete(rw, r)
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
	queryDir = path.Join(queryDir)

	localDir := path.Join(config.PublicDirectoryRoot, queryDir)
	if !utils.IsPathValid(localDir) {
		log.Warnf("invalid query, %s", localDir)
		http.Error(rw, "invalid query", 404)
		return
	}

	_, err := os.Stat(path.Join(config.PublicDirectoryRoot, queryDir))
	if err != nil {
		log.Infof("Directory %s does not exit", localDir)
		http.Error(rw, "Directory does not exit", 404)
		return
	}

	metadataList, err := fs.GetFileMetadataList(localDir)
	if err != nil {
		log.Errorf("failed to list download files, err: %v", err)
		http.Error(rw, "Directory does not exit", 404)
	}

	res := &ListFileResponse{
		QueryFolder:  queryDir,
		MetadataList: metadataList,
	}
	res.ToJSON(rw)
	log.Infof("response value - %s, %d", res.QueryFolder, len(res.MetadataList))
}

type ListFileResponse struct {
	QueryFolder  string             `json:"queryFolder"`
	MetadataList []*fs.FileMetadata `json:"metadatas"`
}

func (p *ListFileResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}

func (hdl *ListHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	log.Debug("handle create folder request")

	keys, ok := r.URL.Query()["key"]
	queryDir := ""
	if ok && len(keys[0]) > 0 {
		queryDir = keys[0]
	}
	queryDir = path.Join(queryDir)

	newDir := path.Join(config.PublicDirectoryRoot, queryDir)
	if !utils.IsPathValid(newDir) {
		log.Warnf("invalid query, %s", newDir)
		http.Error(rw, "invalid query", 404)
		return
	}

	err := os.Mkdir(newDir, os.ModePerm)
	if err != nil {
		log.Errorf("Unable to create %s, err: ", queryDir, err)
		http.Error(rw, "Unable to create folder", 404)
		return
	}

	rw.Write([]byte(queryDir))
	log.Infof("directory created, query: %s, path: %s", queryDir, newDir)
}

func (hdl *ListHandler) handleDelete(rw http.ResponseWriter, r *http.Request) {
	log.Debug("handle delete folder request")

	keys, ok := r.URL.Query()["key"]
	queryPath := ""
	if !ok || len(keys[0]) == 0 {
		log.Info("empty query")
		http.Error(rw, "empty query", 404)
		return
	}
	queryPath = path.Join(keys[0])

	localPath := path.Join(config.PublicDirectoryRoot, queryPath)
	if queryPath == "" || !utils.CheckPathForDelete(localPath) {
		log.Warnf("invalid query, %s", queryPath)
		http.Error(rw, "invalid query", 404)
		return
	}

	err := os.RemoveAll(localPath)
	if err != nil {
		log.Errorf("Failed to delete %s, err: ", localPath, err)
		http.Error(rw, "Cannot find object", 404)
		return
	}

	rw.Write([]byte(queryPath))
	log.Infof("deleted query: %s, path: %s", queryPath, localPath)
}
