package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lyokalita/naspublic.ftserver/src/auth"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

func GetTokenFromHeader(authHeader string) (string, error) {
	headerArr := utils.SplitRemoveEmpty(authHeader, ' ')
	if len(headerArr) < 2 {
		return "", fmt.Errorf("error format authetication %s", authHeader)
	}
	if strings.ToLower(headerArr[0]) != "bearer" {
		return "", fmt.Errorf("error format authetication %s", authHeader)
	}
	return headerArr[1], nil
}

// Check existence and validation of token from request
func ValidateJwtAuthorization(rw http.ResponseWriter, r *http.Request) (*auth.FsPermission, error) {
	// Get Jwt token
	authHeader := r.Header.Get("Authorization")
	token, err := GetTokenFromHeader(authHeader)
	if err != nil {
		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return nil, err
	}

	// check token is valid and not expired
	fsPermission, err := auth.ValidateJwtToken(token)
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			http.Error(rw, "Token expired", http.StatusUnauthorized)
		} else {
			http.Error(rw, "Invalid token", http.StatusUnauthorized)
		}
		return nil, err
	}
	return fsPermission, nil
}

func GetQueryParam(param string, r *http.Request) string {
	keys, ok := r.URL.Query()[param]
	key := ""
	if ok && len(keys[0]) > 0 {
		key = keys[0]
	}
	return key
}
