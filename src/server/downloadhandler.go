package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/auth"
	"github.com/lyokalita/naspublic.ftserver/src/fs"
	"github.com/lyokalita/naspublic.ftserver/src/routine"
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
	if r.Method == http.MethodPost {
		hdl.handlePost(rw, r)
		return
	}
	rw.WriteHeader(http.StatusMethodNotAllowed)
}

/*
Download a file

GET /api/nas/v0/download?key={file path}
*/
func (hdl *DownloadHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	log.Debugf("handle file download request, remote: %s", r.RemoteAddr)

	// get query parameter
	signed := GetQueryParam("signed", r)
	nonce := GetQueryParam("nc", r)

	// validate signed key
	metadata, err := auth.DLSigning.Validate(signed, nonce)
	if err != nil {
		log.Error(err)
		http.Error(rw, "Invalid signed key", http.StatusUnauthorized)
		return
	}

	// check file exists
	info, err := os.Stat(metadata.FilePath)
	if err != nil || info.IsDir() {
		log.Errorf("file does not exist, %s, err: %v", metadata.FilePath, err)
		http.Error(rw, "File does not exist", http.StatusNotFound)
		return
	}
	defer func() {
		if metadata.Type == auth.SIGN_ZIPPED {
			routine.CleanFile(metadata.FilePath)
		}
	}()

	// send file
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(metadata.FilePath)))
	http.ServeFile(rw, r, metadata.FilePath)
	log.Infof("file served: %s", metadata.FilePath)
}

func (hdl *DownloadHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	fsPermission, err := ValidateJwtAuthorization(rw, r)
	if err != nil {
		log.Error(err)
		return
	}

	// get request body
	req := &DownloadPostRequest{}
	err = req.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		log.Error(err)
		return
	}

	// check validity of each requested file
	requestedFileList := []string{}
	for _, file := range req.Files {
		fullFilePath, err := fsPermission.CheckRead(file)
		if err != nil {
			log.Errorf("%v, err: %v", *fsPermission, err)
			http.Error(rw, fmt.Sprintf("No permission to %s", file), http.StatusForbidden)
			return
		}
		_, err = os.Stat(fullFilePath)
		if err != nil {
			log.Errorf("file %s does not exit, err: %v", fullFilePath, err)
			http.Error(rw, fmt.Sprintf("File %s does not exit", file), http.StatusNotFound)
			return
		}
		requestedFileList = append(requestedFileList, fullFilePath)
	}

	// obtain download file path
	downloadFilePath := ""
	signType := auth.SIGN_REGULAR
	if len(requestedFileList) == 0 {
		log.Error("empty requested file list")
		return
	}
	info, _ := os.Stat(requestedFileList[0])
	if len(requestedFileList) == 1 && !info.IsDir() { // serve file directly if only one file is requested
		downloadFilePath = requestedFileList[0]
	} else { // zip files first if multiple requested files
		downloadFilePath, err = fs.ServeMultipleFilesWithCompression(requestedFileList)
		signType = auth.SIGN_ZIPPED
		if err != nil {
			log.Error(err)
			http.Error(rw, "Failed to zip files", http.StatusNotFound)
			return
		}
	}

	// generate signing key
	signed, nonce, err := auth.DLSigning.Generate(&auth.SignedMetadata{TokenId: fsPermission.Id(), FilePath: downloadFilePath, ExpAt: fsPermission.ExpAt(), Type: signType})
	if err != nil {
		log.Errorf("failed to sign %s, %s, err: %v", fsPermission.Id(), downloadFilePath, err)
		return
	}
	res := &DownloadPostResponse{
		Signed: signed,
		Nonce:  nonce,
	}
	res.ToJSON(rw)
	log.Infof("signed id: %s, download path: %s, num files: %d, remote: %v", fsPermission.Id(), downloadFilePath, len(requestedFileList), r.RemoteAddr)
}

type DownloadPostRequest struct {
	Files []string `json:"files"`
}

func (p *DownloadPostRequest) FromJSON(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(p)
}

type DownloadPostResponse struct {
	Signed string `json:"signed"`
	Nonce  string `json:"nonce"`
}

func (p *DownloadPostResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}
