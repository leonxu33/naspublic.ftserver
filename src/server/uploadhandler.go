package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/fs"
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
	fsPermission, err := ValidateJwtAuthorization(rw, r)
	if err != nil {
		log.Error(err)
		return
	}

	queryDir := GetQueryParam("key", r)
	queryDir = path.Join(queryDir)

	// check permission and get full directory
	fullQueryPath, err := fsPermission.CheckWrite(queryDir)
	if err != nil {
		log.Errorf("%v, err: %v", *fsPermission, err)
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
	f_in, header, err := r.FormFile("uploadFile")
	if err != nil {
		log.Errorf("failed to retrieve file, err: ", err)
		http.Error(rw, "Invalid file", http.StatusBadRequest)
		return
	}
	defer f_in.Close()
	log.Infof("Upload File: %s, File Size: %v, MIME Header: %v", header.Filename, header.Size, header.Header)

	// check file name length
	if len(header.Filename) > 250 {
		log.Errorf("file name is too long: %d", len(header.Filename))
		http.Error(rw, "File name too long", http.StatusBadRequest)
		return
	}

	// check path valid
	destinationFilePath := path.Join(fullQueryPath, header.Filename)
	if !validate.IsPathInclusive(fullQueryPath, destinationFilePath) {
		log.Errorf("invalid file name, %s", destinationFilePath)
		http.Error(rw, "Invalid file name", http.StatusBadRequest)
		return
	}

	// check file exists
	_, err = os.Stat(destinationFilePath)
	if err == nil {
		log.Errorf("file already exists, %s", destinationFilePath)
		http.Error(rw, "File already exists", http.StatusConflict)
		return
	}

	// save file
	fileWriter := fs.NewFileUploader(f_in, header.Size, partSize, cancelChan)
	go func() {
		defer close(responseChan)
		totalWriteSize, err := fileWriter.WriteTo(destinationFilePath)
		if err != nil {
			log.Errorf("failed to write to file %s, err: %v", destinationFilePath, err)
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
