package routine

import (
	"os"

	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/validate"
)

func CleanFile(filePath string) {
	if validate.IsPathInclusive(config.TempDirectoryRoot, filePath) {
		_ = os.Remove(filePath)
	}
}
