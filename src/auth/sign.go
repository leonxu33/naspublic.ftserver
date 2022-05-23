package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

var DLSigning *Signing = &Signing{
	keyMap: map[string]string{},
}

type Signing struct {
	keyMap map[string]string
}

type SignedMetadata struct {
	TokenId  string
	FilePath string
	ExpAt    int64
	Type     string
}

const SIGN_REGULAR = "regular"
const SIGN_ZIPPED = "zipped"

/*
Generate signing key for the inputs
*/
func (m *Signing) Generate(signedMetadata *SignedMetadata) (string, string, error) {
	metadata := m.encodeSignedMetadata(signedMetadata)
	block, err := aes.NewCipher(config.SignSecret)
	if err != nil {
		return "", "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	nonce := utils.GetRandomBytes(aesgcm.NonceSize())
	if err != nil {
		return "", "", err
	}
	cipherText := aesgcm.Seal(nil, nonce, []byte(metadata), nil)
	signedKey := hex.EncodeToString(cipherText)

	m.keyMap[signedKey] = string(md5.New().Sum([]byte(metadata)))
	return signedKey, hex.EncodeToString(nonce), nil
}

/*
Validate signing key
*/
func (m *Signing) Validate(signedKey string, nonce string) (*SignedMetadata, error) {
	cipherText, err := hex.DecodeString(signedKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(config.SignSecret)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceHexDecoded, err := hex.DecodeString(nonce)
	if err != nil {
		return nil, err
	}
	metadataInByte, err := aesgcm.Open(nil, nonceHexDecoded, cipherText, nil)
	if err != nil {
		return nil, err
	}

	if chksum, ok := m.keyMap[signedKey]; ok {
		if chksum != string(md5.New().Sum(metadataInByte)) {
			return nil, fmt.Errorf("signing key not correct")
		}
	} else {
		return nil, fmt.Errorf("signing key not found")
	}
	delete(m.keyMap, signedKey)

	metadataInString := string(metadataInByte)
	metadata, err := m.decodeSignedMetadata(metadataInString)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (m *Signing) encodeSignedMetadata(signedMetadata *SignedMetadata) string {
	return fmt.Sprintf("%s,%s,%v,%s", signedMetadata.TokenId, signedMetadata.FilePath, signedMetadata.ExpAt, signedMetadata.Type)
}

func (m *Signing) decodeSignedMetadata(encodedString string) (*SignedMetadata, error) {
	arr := utils.SplitRemoveEmpty(encodedString, ',')
	if len(arr) != 4 {
		return nil, fmt.Errorf("error metadata: %s", encodedString)
	}

	expAt, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil || expAt <= time.Now().Unix() {
		return nil, fmt.Errorf("expired: %s", utils.ConvertUnixTimeToString(expAt))
	}

	return &SignedMetadata{
		TokenId:  arr[0],
		FilePath: arr[1],
		ExpAt:    expAt,
		Type:     arr[3],
	}, nil
}
