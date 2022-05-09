package filebrowser

import (
	"io/ioutil"
	"path"

	"github.com/lyokalita/naspublic.ftserver/config"
)

type FileMetadata struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	LastModified string `json:"date"`
}

func GetFileMetadataList(dirPath string) ([]*FileMetadata, error) {
	files, err := ioutil.ReadDir(path.Join(config.PublicDirectoryRoot, dirPath))
	if err != nil {
		return nil, err
	}
	metadataList := []*FileMetadata{}
	for _, file := range files {
		fileType := ""
		if file.IsDir() {
			fileType = "Folder"
		}
		metadataList = append(metadataList, &FileMetadata{
			Name:         file.Name(),
			Size:         file.Size(),
			Type:         fileType,
			LastModified: file.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	return metadataList, err
}
