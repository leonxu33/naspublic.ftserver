package fs

import (
	"bufio"
	"fmt"
	"io"
	"os"

	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

type FileUploader struct {
	FileSize   int64
	PartSize   int64
	ReaderAt   io.ReaderAt
	CancelChan <-chan int
}

func NewFileUploader(inputFileReader io.ReaderAt, fileSize int64, partSize int64, cancelChannel <-chan int) *FileUploader {
	return &FileUploader{
		ReaderAt:   inputFileReader,
		FileSize:   fileSize,
		PartSize:   partSize,
		CancelChan: cancelChannel,
	}
}

func (fw *FileUploader) WriteTo(destinationPath string) (int64, error) {
	err := fw.setupDestinationFile(destinationPath)
	if err != nil {
		return 0, err
	}

	f_out, err := os.OpenFile(destinationPath, os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f_out.Close()

	totalWriteSize, err := fw.writeFile(f_out)
	if err != nil {
		return totalWriteSize, err
	}

	if totalWriteSize != fw.FileSize {
		return totalWriteSize, fmt.Errorf("final write size %d does not match inital file size %d", totalWriteSize, fw.FileSize)
	}

	log.Debug(utils.PrintMemUsage())
	return totalWriteSize, nil
}

func (fw *FileUploader) setupDestinationFile(destinationPath string) error {
	f, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	f.Close()
	err = os.Truncate(destinationPath, fw.FileSize)
	if err != nil {
		return err
	}
	return nil
}

func (fw *FileUploader) writeFile(w io.Writer) (int64, error) {
	bufferedWriter := bufio.NewWriter(w)
	var totalWriteSize int64 = 0
	var offset int64 = 0
	for ; offset < fw.FileSize; offset += fw.PartSize {
		select {
		case <-fw.CancelChan:
			return totalWriteSize, fmt.Errorf("file writer is canceled")
		default:
			n, err := fw.writePart(bufferedWriter, offset)
			if err != nil {
				return totalWriteSize, err
			}
			totalWriteSize += int64(n)
		}
	}
	err := bufferedWriter.Flush()
	if err != nil {
		return totalWriteSize, err
	}
	return totalWriteSize, err
}

func (fw *FileUploader) writePart(w *bufio.Writer, offset int64) (int, error) {
	buffer := make([]byte, fw.PartSize)
	readSize, err := fw.ReaderAt.ReadAt(buffer, offset)
	if err != nil && err.Error() != "EOF" {
		return 0, err
	}
	if readSize != int(fw.PartSize) {
		buffer = buffer[:readSize]
	}
	writeSize, err := w.Write(buffer)
	if err != nil {
		return writeSize, err
	}
	if readSize != writeSize {
		return writeSize, fmt.Errorf("write size %d does not match read size %d", writeSize, readSize)
	}
	return writeSize, nil
}
