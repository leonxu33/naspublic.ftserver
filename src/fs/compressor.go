package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/lyokalita/naspublic.ftserver/src/config"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

func ServeMultipleFilesWithCompression(fileList []string) (string, error) {
	// 1. Create a ZIP file and zip.Writer
	zipName := fmt.Sprintf("%s_%s.zip", utils.GetCurrentTimeCompact(), string(utils.GetRandomBytes(8)))
	zipTarget := path.Join(config.TempDirectoryRoot, zipName)
	f, err := os.Create(zipTarget)
	if err != nil {
		return "", err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	for _, file := range fileList {
		err := zipFile(file, writer)
		if err != nil {
			return "", err
		}
	}
	return zipTarget, nil
}

func zipFile(source string, writer *zip.Writer) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
