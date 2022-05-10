package utils

import (
	"fmt"
	"math/rand"
	"path"
	"runtime"
	"strings"

	"github.com/lyokalita/naspublic.ftserver/src/config"
)

func GetFileWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(path.Ext(fileName))]
}

func GetUniqueFileName(fileName string) string {
	return fmt.Sprintf("%s_%s%s", fileName[:len(fileName)-len(path.Ext(fileName))], randSeq(6), path.Ext(fileName))
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func PrintMemUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Mem usage: Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v\n", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
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

func IsPathValid(queryPath string) bool {
	publicRoot := path.Join(config.PublicDirectoryRoot)
	//	log.Infof("%s, %s", publicRoot, queryPath)

	if len(queryPath) < len(publicRoot) { // query path must be no shorter than public root
		return false
	}

	if queryPath[:len(publicRoot)] != publicRoot { // query path must start with public root
		return false
	}

	return true
}

func CheckPathForDelete(queryPath string) bool {
	if !IsPathValid(queryPath) {
		return false
	}

	publicRoot := path.Join(config.PublicDirectoryRoot)

	if publicRoot == queryPath { // for delete, query path cannot be the same as public root
		return false
	}

	return true
}
