package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"path"
	"runtime"
	"strings"
)

func GetFileWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(path.Ext(fileName))]
}

func GetUniqueFileName(fileName string) string {
	return fmt.Sprintf("%s_%s%s", fileName[:len(fileName)-len(path.Ext(fileName))], GetRandomBytes(6), path.Ext(fileName))
}

var letters = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GetRandomBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[GetRandomNumber(int64(len(letters)))]
	}
	return b
}

func GetRandomNumber(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}

func PrintMemUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Mem usage: Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v\n", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func SplitRemoveEmpty(str string, token rune) []string {
	f := func(c rune) bool {
		return c == token
	}
	return strings.FieldsFunc(str, f)
}

// Return extension(without dot) if path is a file.
func GetFileExtension(filePath string) string {
	if filePath == "" {
		return ""
	}
	ext := path.Ext(filePath)
	if ext == "" {
		return ""
	}
	return strings.Split(ext, ".")[1]
}

func GetTokenFromHeader(authHeader string) (string, error) {
	headerArr := SplitRemoveEmpty(authHeader, ' ')
	if len(headerArr) < 2 {
		return "", fmt.Errorf("Error format authorization %s", authHeader)
	}
	if strings.ToLower(headerArr[0]) != "bearer" {
		return "", fmt.Errorf("Error format authorization %s", authHeader)
	}
	return headerArr[1], nil
}
