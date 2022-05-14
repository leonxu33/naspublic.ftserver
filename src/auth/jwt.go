package auth

import (
	"fmt"
	"path"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
	"github.com/lyokalita/naspublic.ftserver/src/validate"
)

type VisitorAuthField struct {
	Mode string
	Dir  string
}

type VisitorClaim struct {
	*jwt.StandardClaims
	TokenType string
	VisitorAuthField
}

/*
param:
- mode: access permission, r: read/download, w: write/upload, d: delete;
- dir: allowed directory;
- valid: valid period in minutes, expiration date = token creation date + valid;

return:
- signed token string
*/
func GenerateJwtToken(mode string, dir string, valid int64) (string, error) {
	expAt := time.Now().Add(time.Minute * time.Duration(valid)).Unix()

	t := jwt.New(jwt.GetSigningMethod("HS256"))
	t.Claims = &VisitorClaim{
		&jwt.StandardClaims{
			Id:        string(utils.GetRandomBytes(8)),
			ExpiresAt: expAt,
		},
		"level1",
		VisitorAuthField{
			Mode: mode,
			Dir:  dir,
		},
	}

	return t.SignedString(config.JwtSecret)
}

func ValidateJwtToken(tokenString string) (*FsPermission, error) {
	token, err := jwt.ParseWithClaims(tokenString, &VisitorClaim{}, func(token *jwt.Token) (interface{}, error) {
		return config.JwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*VisitorClaim)

	if !validate.IsModeValid(claims.Mode) {
		return nil, fmt.Errorf("invalid permission mode")
	}

	completeDir := path.Join(config.PublicDirectoryRoot, claims.Dir)
	if !validate.IsPathInclusive(config.PublicDirectoryRoot, completeDir) {
		return nil, fmt.Errorf("invalid permission directory")
	}

	return &FsPermission{
		id:        claims.Id,
		read:      claims.Mode[0] == validate.READ_MODE,
		write:     claims.Mode[1] == validate.WRITE_MODE,
		delete:    claims.Mode[2] == validate.DELETE_MODE,
		directory: completeDir,
	}, nil
}
