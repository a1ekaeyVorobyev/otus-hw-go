package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrUnsupportedFile       = errors.New("unsupported file")
)

func Copy(fromPath string, toPath string, offset, limit int64) error {
	// Place your code here
	var reader, barReader io.Reader
	fileSource, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fileSource.Close()

	fileDestination, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer fileDestination.Close()

	fs, _ := fileSource.Stat()
	if fs.Size() <= offset {
		return ErrOffsetExceedsFileSize
	}
	if !fs.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if limit == 0 || limit+offset > fs.Size() {
		limit = fs.Size() - offset
	}

	if o, err := fileSource.Seek(offset, 0); err != nil || o != offset {
		return err
	}

	fmt.Printf("Coping file %v to file %v\n", fromPath, toPath)
	bar := pb.Full.Start64(limit)
	reader = io.LimitReader(fileSource, limit)
	barReader = bar.NewProxyReader(reader)
	_, err = io.Copy(fileDestination, barReader)
	if err != nil {
		return err
	}
	bar.Finish()
	return nil
}
