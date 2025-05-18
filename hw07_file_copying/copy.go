//nolint:revive
package main

import (
	"errors"
	"io"
	"os"
	"strings"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file") //nolint:revive
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrUnsupportedPath       = errors.New("unsupported path")
)

var unsupportedPathPrefixes = []string{
	"/dev",
	"/proc",
	"/sys",
	"/run",
	"/var/run",
	"/var/lock",
}

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

func isSystemPath(path string) bool {
	for _, prefix := range unsupportedPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func isInvalidToPath(path string) bool {
	res := isSystemPath(path)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return true
	}
	return res || fileInfo.IsDir()
}

//nolint:revive
func Copy(fromPath, toPath string, offset, limit int64) error {
	if isUnsupportedFile(fromPath) {
		return ErrUnsupportedFile
	}

	if isSystemPath(fromPath) {
		return ErrUnsupportedPath
	}

	if isInvalidToPath(toPath) {
		return ErrUnsupportedPath
	}

	return nil
}
