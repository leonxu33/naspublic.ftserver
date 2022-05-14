package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/auth"
	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/validate"
)

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (hdl *AuthHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		hdl.handlePost(rw, r)
		return
	}
	if r.Method == http.MethodGet {
		hdl.handleGet(rw, r)
		return
	}
	rw.WriteHeader(http.StatusMethodNotAllowed)
}

func (hdl *AuthHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	req := &TokenRequest{}
	err := req.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Invalid input", http.StatusBadRequest)
		log.Info(err)
		return
	}

	err = req.Validate()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		log.Info(err)
		return
	}

	tokenString, err := auth.GenerateJwtToken(req.Mode, req.Dir, req.Valid)
	if err != nil {
		http.Error(rw, "Failed to create token", http.StatusBadRequest)
		log.Info(err)
		return
	}
	rw.Write([]byte(tokenString))
	log.Infof("created new token for %v: %s", *req, tokenString)
}

type TokenRequest struct {
	Mode  string `json:"mode"`
	Dir   string `json:"dir"`
	Valid int64  `json:"valid"`
}

func (p *TokenRequest) FromJSON(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(p)
}

func (p *TokenRequest) Validate() error {
	// validate mode
	if !validate.IsModeValid(p.Mode) {
		return fmt.Errorf("invalid permission mode")
	}

	// validate dir
	completePath := path.Join(config.PublicDirectoryRoot, p.Dir)
	if !validate.IsPathInclusive(config.PublicDirectoryRoot, completePath) {
		return fmt.Errorf("invalid permission directory")
	}

	// validate valid
	if p.Valid <= 0 {
		return fmt.Errorf("invalid expiration period")
	}
	return nil
}

func (hdl *AuthHandler) handleGet(rw http.ResponseWriter, r *http.Request) {
	input, err := io.ReadAll(r.Body)
	if err != nil {
		log.Info(err)
		return
	}

	tokenString := string(input)
	log.Info(tokenString)
	permission, err := auth.ValidateJwtToken(tokenString)
	if err != nil {
		log.Info(err)
		return
	}
	log.Infof("%v", *permission)
}