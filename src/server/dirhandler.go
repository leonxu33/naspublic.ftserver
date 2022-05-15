package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/auth"
	"github.com/lyokalita/naspublic.ftserver/src/fs"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

type DirHandler struct {
}

func NewDirHandler() *DirHandler {
	return &DirHandler{}
}

func (hdl *DirHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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

/*
List all files and folders in a directory

GET /api/nas/v0/dir?key={directory path}
*/
func (hdl *DirHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	// Get Jwt token
	authHeader := r.Header.Get("Authorization")
	token, err := utils.GetTokenFromHeader(authHeader)
	if err != nil {
		log.Info(err)
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}

	// check token is valid and not expired
	fsPermission, err := auth.ValidateJwtToken(token)
	if err != nil {
		log.Info(err)
		if strings.Contains(err.Error(), "expired") {
			http.Error(rw, "Token expired", http.StatusUnauthorized)
		} else {
			http.Error(rw, "Invalid token", http.StatusUnauthorized)
		}
		return
	}

	// get query parameter
	keys, ok := r.URL.Query()["key"]
	queryDir := ""
	if ok && len(keys[0]) > 0 {
		queryDir = keys[0]
	}
	queryDir = path.Join(queryDir)

	// check permission and get full directory
	fullQueryPath, err := fsPermission.CheckRead(queryDir)
	if err != nil {
		log.Infof("%v, err: %v", *fsPermission, err)
		http.Error(rw, "No permission", http.StatusForbidden)
		return
	}

	// check directory exists
	_, err = os.Stat(fullQueryPath)
	if err != nil {
		log.Infof("directory %s does not exit, err: %v", fullQueryPath, err)
		http.Error(rw, "Directory does not exit", http.StatusNotFound)
		return
	}

	// get file list in the directory
	metadataList, err := fs.GetFileMetadataList(fullQueryPath)
	if err != nil {
		log.Errorf("failed to list files, err: %v", err)
		http.Error(rw, "Directory does not exit", http.StatusNotFound)
		return
	}

	res := &ListFileResponse{
		QueryFolder:  queryDir,
		MetadataList: metadataList,
	}
	res.ToJSON(rw)
	log.Infof("list full path: %s, query path: %s, num files: %d, remote: %s", fullQueryPath, res.QueryFolder, len(res.MetadataList), r.RemoteAddr)
}

type ListFileResponse struct {
	QueryFolder  string             `json:"queryFolder"`
	MetadataList []*fs.FileMetadata `json:"metadatas"`
}

func (p *ListFileResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}

/*
Create a directory

POST /api/nas/v0/dir?key={directory path}
*/
func (hdl *DirHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	// Get Jwt token
	authHeader := r.Header.Get("Authorization")
	token, err := utils.GetTokenFromHeader(authHeader)
	if err != nil {
		log.Info(err)
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}

	// check token is valid and not expired
	fsPermission, err := auth.ValidateJwtToken(token)
	if err != nil {
		log.Info(err)
		if strings.Contains(err.Error(), "expired") {
			http.Error(rw, "Token expired", http.StatusUnauthorized)
		} else {
			http.Error(rw, "Invalid token", http.StatusUnauthorized)
		}
		return
	}

	// get query parameter
	keys, ok := r.URL.Query()["key"]
	queryDir := ""
	if ok && len(keys[0]) > 0 {
		queryDir = keys[0]
	}
	queryDir = path.Join(queryDir)

	// check permission and get full directory
	fullQueryPath, err := fsPermission.CheckWrite(queryDir)
	if err != nil {
		log.Infof("%v, err: %v", *fsPermission, err)
		http.Error(rw, "No permission", http.StatusForbidden)
		return
	}

	// check directory exists
	_, err = os.Stat(fullQueryPath)
	if err == nil {
		log.Infof("directory %s already exists", fullQueryPath)
		http.Error(rw, "Directory already exists", http.StatusConflict)
		return
	}

	// create a new folder
	err = os.Mkdir(fullQueryPath, os.ModePerm)
	if err != nil {
		log.Errorf("Unable to create %s, err: ", queryDir, err)
		http.Error(rw, "Unable to create folder", http.StatusNotFound)
		return
	}

	rw.Write([]byte(queryDir))
	log.Infof("directory created, query: %s, path: %s, remote: %s", queryDir, fullQueryPath, r.RemoteAddr)
}

/*
Delete a target

DELETE /api/nas/v0/dir?key={target path}
*/
func (hdl *DirHandler) handleDelete(rw http.ResponseWriter, r *http.Request) {
	// Get Jwt token
	authHeader := r.Header.Get("Authorization")
	token, err := utils.GetTokenFromHeader(authHeader)
	if err != nil {
		log.Info(err)
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}

	// check token is valid and not expired
	fsPermission, err := auth.ValidateJwtToken(token)
	if err != nil {
		log.Info(err)
		if strings.Contains(err.Error(), "expired") {
			http.Error(rw, "Token expired", http.StatusUnauthorized)
		} else {
			http.Error(rw, "Invalid token", http.StatusUnauthorized)
		}
		return
	}

	// get query parameter
	keys, ok := r.URL.Query()["key"]
	queryPath := ""
	if ok && len(keys[0]) > 0 {
		queryPath = keys[0]
	}
	queryPath = path.Join(queryPath)

	// check permission and get full directory
	fullQueryPath, err := fsPermission.CheckDelete(queryPath)
	if err != nil {
		log.Infof("%v, err: %v", *fsPermission, err)
		http.Error(rw, "No permission", http.StatusForbidden)
		return
	}

	// check path exists
	_, err = os.Stat(fullQueryPath)
	if err != nil {
		log.Infof("path %s does not exit, err: %v", fullQueryPath, err)
		http.Error(rw, "Target not exist", http.StatusNotFound)
		return
	}

	err = os.RemoveAll(fullQueryPath)
	if err != nil {
		log.Errorf("failed to delete %s, err: ", fullQueryPath, err)
		http.Error(rw, "Cannot find object", http.StatusNotFound)
		return
	}

	rw.Write([]byte(queryPath))
	log.Infof("deleted query: %s, path: %s, remote: %s", queryPath, fullQueryPath, r.RemoteAddr)
}
