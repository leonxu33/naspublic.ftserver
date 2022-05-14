package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/auth"
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
	log.Debug("handle file download request")

	// get query parameter
	signedParam, ok := r.URL.Query()["signed"]
	signed := ""
	if ok && len(signedParam[0]) > 0 {
		signed = signedParam[0]
	}

	nonceParam, ok := r.URL.Query()["nc"]
	nonce := ""
	if ok && len(nonceParam[0]) > 0 {
		nonce = nonceParam[0]
	}

	// validate signed key
	_, fullQueryPath, err := auth.DLSigning.Validate(signed, nonce)
	if err != nil {
		log.Infof("invalid signed key")
		http.Error(rw, "Invalid signed key", http.StatusUnauthorized)
		return
	}

	// check file exists
	info, err := os.Stat(fullQueryPath)
	if err != nil || info.IsDir() {
		log.Infof("file does not exist, %s, err: %v", fullQueryPath, err)
		http.Error(rw, "File does not exist", http.StatusNotFound)
		return
	}

	// send file
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(fullQueryPath)))
	http.ServeFile(rw, r, fullQueryPath)
	log.Infof("file %s served", fullQueryPath)
}

func (hdl *DownloadHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
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
	queryFile := ""
	if ok && len(keys[0]) > 0 {
		queryFile = keys[0]
	}
	queryFile = path.Join(queryFile)

	// check permission and get full directory
	fullQueryPath, err := fsPermission.CheckRead(queryFile)
	if err != nil {
		log.Infof("%v, err: %v", *fsPermission, err)
		http.Error(rw, "No permission", http.StatusForbidden)
		return
	}

	// generate signing key
	signed, nonce, err := auth.DLSigning.Generate(fsPermission.GetId(), fullQueryPath)
	if err != nil {
		log.Infof("failed to sign %s, %s, err: %v", fsPermission.GetId(), fullQueryPath, err)
		return
	}
	res := &SigningKeyResponse{
		Signed: signed,
		Nonce:  nonce,
	}
	res.ToJSON(rw)
	log.Infof("signed %s, %s, %v", fsPermission.GetId(), fullQueryPath, res)
}

type SigningKeyResponse struct {
	Signed string `json:"signed"`
	Nonce  string `json:"nonce"`
}

func (p *SigningKeyResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}
