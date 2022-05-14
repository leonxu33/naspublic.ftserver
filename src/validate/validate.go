package validate

func CheckPathForDelete(parentPath string, childPath string) bool {
	if !IsPathInclusive(parentPath, childPath) {
		return false
	}

	if parentPath == childPath { // for delete, child path cannot be the same as parent root
		return false
	}

	return true
}

func IsPathInclusive(parentPath string, childPath string) bool {
	if len(childPath) < len(parentPath) { // child path must be no shorter than parent path
		return false
	}

	if childPath[:len(parentPath)] != parentPath { // child path must start with parent root
		return false
	}
	return true
}

const (
	READ_MODE    = 'r'
	WRITE_MODE   = 'w'
	DELETE_MODE  = 'd'
	DISABLE_RUNE = '-'
)

var MODE_RUNES []byte = []byte{READ_MODE, WRITE_MODE, DELETE_MODE}

func IsModeValid(mode string) bool {
	if len(mode) != len(MODE_RUNES) {
		return false
	}
	for i, m := range MODE_RUNES {
		if mode[i] != m && mode[i] != DISABLE_RUNE {
			return false
		}
	}
	return true
}
