//nolint:revive
package main

import (
	"errors"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file") //nolint:revive
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrUnsupportedPath       = errors.New("unsupported path")
)

func isUnsupportedFile(fromPath string) bool {
	fileInfo, err := os.Lstat(fromPath)
	// File not exists.
	if err != nil {
		return true
	}

	// Not a regular file or a symlink.
	if !fileInfo.Mode().IsRegular() || (fileInfo.Mode()&os.ModeSymlink != 0) {
		return true
	}

	// Unknown file size.
	f, err := os.Open(fromPath)
	if err != nil {
		return true
	}
	defer f.Close()
	_, err = f.Seek(0, io.SeekEnd)

	return err != nil
}

//nolint:revive
func Copy(fromPath, toPath string, offset, limit int64) error {
	if isUnsupportedFile(fromPath) {
		return ErrUnsupportedFile
	}

	// fmt.Println(fromPath, toPath, offset, limit)
	// stat, err := os.Stat(fromPath)
	// fmt.Println(stat.Mode().IsDir(), stat.Mode().IsRegular(), stat.Size())
	// if err != nil {
	// 	return ErrUnsupportedFile
	// }
	// fmt.Println(stat)
	return nil
}
