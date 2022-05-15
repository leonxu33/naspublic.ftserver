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
	"github.com/lyokalita/naspublic.ftserver/src/validate"
)

type UploadHandler struct {
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (hdl *UploadHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		hdl.handlePost(rw, r)
		return
	}
	rw.WriteHeader(http.StatusMethodNotAllowed)
}

/*
Upload a file

POST /api/nas/v0/upload?key={file path}
*/
func (hdl *UploadHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
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

	// begin progress remote data
	log.Debugf("handle file upload request full path: %s, remote: %s", fullQueryPath, r.RemoteAddr)
	ctx := r.Context()
	cancelChan := make(chan int)
	responseChan := make(chan *UploadResponse)
	var partSize int64 = 10 << 20

	// fetch remote data
	r.ParseMultipartForm(partSize)
	f_in, handler, err := r.FormFile("uploadFile")
	if err != nil {
		log.Errorf("failed to retrieve file, err: ", err)
		http.Error(rw, "Invalid file", http.StatusBadRequest)
		return
	}
	defer f_in.Close()
	log.Infof("Upload File: %s, File Size: %v, MIME Header: %v", handler.Filename, handler.Size, handler.Header)

	// check file name length
	if len(handler.Filename) > 250 {
		log.Infof("file name is too long: %d", len(handler.Filename))
		http.Error(rw, "File name too long", http.StatusBadRequest)
		return
	}

	// check path valid
	destinationFilePath := path.Join(fullQueryPath, handler.Filename)
	if !validate.IsPathInclusive(fullQueryPath, destinationFilePath) {
		log.Infof("invalid file name, %s", destinationFilePath)
		http.Error(rw, "Invalid file name", http.StatusBadRequest)
		return
	}

	// check file exists
	_, err = os.Stat(destinationFilePath)
	if err == nil {
		log.Infof("file already exists, %s", destinationFilePath)
		http.Error(rw, "File already exists", http.StatusConflict)
		return
	}

	// save file
	fileWriter := fs.NewFileUploader(f_in, handler.Size, partSize, cancelChan)
	go func() {
		defer close(responseChan)
		totalWriteSize, err := fileWriter.WriteTo(destinationFilePath)
		if err != nil {
			log.Infof("failed to write to file %s, err: %v", destinationFilePath, err)
			err = os.Remove(destinationFilePath)
			if err != nil {
				log.Errorf("failed to clean file %s", destinationFilePath)
			}
			responseChan <- &UploadResponse{
				Status:  -1,
				Message: "unable to upload file",
			}
		} else {
			log.Infof("successfully wrote to file %s with %d bytes", destinationFilePath, totalWriteSize)
			responseChan <- &UploadResponse{
				Status:  0,
				Message: "success",
			}
		}
	}()

	select {
	case writerResponse := <-responseChan:
		log.Infof("response value - success: %d, message: %s", writerResponse.Status, writerResponse.Message)
		if writerResponse.Status == 0 {
			rw.Write([]byte("Upload successfully"))
		} else {
			http.Error(rw, "Upload failed", http.StatusNotFound)
		}
	case <-ctx.Done():
		cancelChan <- 1
		log.Infof("request is cancelled, %s", destinationFilePath)
	}
}

type UploadResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (p *UploadResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}
