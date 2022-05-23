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
	"github.com/lyokalita/naspublic.ftserver/src/utils"
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
	authHeader := r.Header.Get("Authorization")
	token, err := GetTokenFromHeader(authHeader)
	if err != nil {
		log.Error(err)
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}
	if token != config.AuthSecret {
		log.Error("token not correct")
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}

	req := &TokenRequest{}
	err = req.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Invalid input", http.StatusBadRequest)
		log.Error(err)
		return
	}

	err = req.Validate()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		log.Error(err)
		return
	}

	tokenString, err := auth.GenerateJwtToken(req.Mode, req.Dir, req.Valid)
	if err != nil {
		http.Error(rw, "Failed to create token", http.StatusBadRequest)
		log.Error(err)
		return
	}
	rw.Write([]byte(tokenString))
	log.Infof("created new token for %v, remote: %s", *req, r.RemoteAddr)
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
	fsPermission, err := ValidateJwtAuthorization(rw, r)
	if err != nil {
		log.Error(err)
		return
	}

	res := &AuthGetResponse{
		Read:   fsPermission.AllowRead(),
		Write:  fsPermission.AllowWrite(),
		Delete: fsPermission.AllowDelete(),
		ExpAt:  utils.ConvertUnixTimeToString(fsPermission.ExpAt()),
	}
	res.ToJSON(rw)
	log.Info(fsPermission.String())
}

type AuthGetResponse struct {
	Read   bool
	Write  bool
	Delete bool
	ExpAt  string
}

func (p *AuthGetResponse) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(p)
}
