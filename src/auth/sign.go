package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
	"github.com/lyokalita/naspublic.ftserver/src/validate"
)

var DLSigning *Signing = &Signing{
	keyMap: map[string]string{},
}

type Signing struct {
	keyMap map[string]string
}

/*
Generate signing key for the inputs

return: signed key, nonce
*/
func (m *Signing) Generate(tokenId string, filePath string) (string, string, error) {
	metadata := fmt.Sprintf("%s,%s", tokenId, filePath)
	block, err := aes.NewCipher(config.SignSecret)
	if err != nil {
		return "", "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	nonce := utils.GetRandomBytes(aesgcm.NonceSize())
	if err != nil {
		return "", "", err
	}
	cipherText := aesgcm.Seal(nil, []byte(nonce), []byte(metadata), nil)
	signedKey := hex.EncodeToString(cipherText)

	m.keyMap[signedKey] = string(md5.New().Sum([]byte(metadata)))
	return signedKey, hex.EncodeToString(nonce), nil
}

func (m *Signing) Validate(signedKey string, nonce string) (string, string, error) {
	cipherText, err := hex.DecodeString(signedKey)
	if err != nil {
		return "", "", err
	}

	block, err := aes.NewCipher(config.SignSecret)
	if err != nil {
		return "", "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonceHexDecoded, err := hex.DecodeString(nonce)
	if err != nil {
		return "", "", err
	}
	metadataInByte, err := aesgcm.Open(nil, nonceHexDecoded, cipherText, nil)
	if err != nil {
		return "", "", err
	}

	if chksum, ok := m.keyMap[signedKey]; ok {
		if chksum != string(md5.New().Sum(metadataInByte)) {
			return "", "", fmt.Errorf("signing key not correct")
		}
	} else {
		return "", "", fmt.Errorf("signing key not found")
	}
	delete(m.keyMap, signedKey)

	metadataInString := string(metadataInByte)
	arr := utils.SplitRemoveEmpty(metadataInString, ',')
	if len(arr) < 2 {
		return "", "", fmt.Errorf("error metadata: %s", metadataInString)
	}

	filePath := arr[1]
	if !validate.IsPathInclusive(config.PublicDirectoryRoot, filePath) {
		return "", "", fmt.Errorf("error file path: %s", filePath)
	}
	return arr[0], filePath, nil
}
