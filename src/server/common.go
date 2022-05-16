package server

import (
	"fmt"
	"strings"

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
