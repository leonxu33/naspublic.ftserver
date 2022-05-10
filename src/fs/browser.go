package fs

import (
	"io/ioutil"

	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

type FileMetadata struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	LastModified string `json:"date"`
}

func GetFileMetadataList(dirPath string) ([]*FileMetadata, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	metadataList := []*FileMetadata{}
	for _, file := range files {
		fileType := ""
		if file.IsDir() {
			fileType = "Folder"
		} else {
			fileType = utils.GetFileExtension(file.Name())
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
