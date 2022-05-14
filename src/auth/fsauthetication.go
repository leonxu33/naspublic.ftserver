package auth

import (
	"fmt"
	"path"

	"github.com/lyokalita/naspublic.ftserver/src/validate"
)

type FsPermission struct {
	id        string
	read      bool
	write     bool
	delete    bool
	directory string
}

func (c *FsPermission) GetId() string {
	return c.id
}

func (c *FsPermission) CheckRead(targetPath string) (string, error) {
	if !c.read {
		return "", fmt.Errorf("no read permission")
	}
	fullTargetPath := path.Join(c.directory, targetPath)
	if !validate.IsPathInclusive(c.directory, fullTargetPath) {
		return "", fmt.Errorf("no read permission to %s", fullTargetPath)
	}
	return fullTargetPath, nil
}

func (c *FsPermission) CheckWrite(targetPath string) (string, error) {
	if !c.write {
		return "", fmt.Errorf("no write permission")
	}
	fullTargetPath := path.Join(c.directory, targetPath)
	if !validate.IsPathInclusive(c.directory, fullTargetPath) {
		return "", fmt.Errorf("no write permission to %s", fullTargetPath)
	}
	return fullTargetPath, nil
}

func (c *FsPermission) CheckDelete(targetPath string) (string, error) {
	if !c.delete {
		return "", fmt.Errorf("no delete permission")
	}
	fullTargetPath := path.Join(c.directory, targetPath)
	if !validate.CheckPathForDelete(c.directory, fullTargetPath) {
		return "", fmt.Errorf("no delete permission to %s", fullTargetPath)
	}
	return fullTargetPath, nil
}
