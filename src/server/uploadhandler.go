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

func (hdl *UploadHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	log.Debug("handle file upload request")

	keys, ok := r.URL.Query()["key"]
	queryDir := ""
	if ok && len(keys[0]) > 0 {
		queryDir = keys[0]
	}
	queryDir = path.Join(queryDir)

	ctx := r.Context()
	cancelChan := make(chan int)
	responseChan := make(chan *UploadResponse)
	var partSize int64 = 10 << 20
	r.ParseMultipartForm(partSize)
	f_in, handler, err := r.FormFile("uploadFile")
	if err != nil {
		log.Errorf("failed to retrieve file, err: ", err)
		res := &UploadResponse{
			Status:  -1,
			Message: "unable to upload file",
		}
		res.ToJSON(rw)
		return
	}
	defer f_in.Close()
	log.Infof("Upload File: %s, File Size: %v, MIME Header: %v", handler.Filename, handler.Size, handler.Header)

	destinationFileName := utils.GetUniqueFileName(handler.Filename)
	if len(destinationFileName) > 250 {
		log.Errorf("file name is too long: %d", len(destinationFileName))
		res := &UploadResponse{
			Status:  -1,
			Message: "file name is too long",
		}
		res.ToJSON(rw)
		return
	}

	destinationFilePath := path.Join(config.PublicDirectoryRoot, queryDir, destinationFileName)
	if !utils.IsPathValid(destinationFilePath) {
		log.Warnf("invalid query, %s", destinationFilePath)
		http.Error(rw, "invalid query", 404)
		return
	}
	fileWriter := fs.NewFileUploader(f_in, handler.Size, partSize, cancelChan)
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
		writerResponse.ToJSON(rw)
		log.Infof("response value - success: %d, message: %s", writerResponse.Status, writerResponse.Message)
	case <-ctx.Done():
		cancelChan <- 1
		log.Info("request is cancelled")
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
