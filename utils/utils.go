package utils

import (
	"fmt"
	"math/rand"
	"path"
	"runtime"
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
