//nolint:revive
package main

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3" //nolint:depguard
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file") //nolint:revive
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrUnsupportedPath       = errors.New("unsupported path")
	ErrOSError               = errors.New("os error")
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

func argCorrection(val *int64) {
	if *val < 0 {
		*val = 0
	}
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

	argCorrection(&offset)
	argCorrection(&limit)

	sourceFile, _ := os.Open(fromPath)
	defer sourceFile.Close()

	// Offset validation.
	fileInfo, _ := os.Stat(fromPath)
	if offset >= fileInfo.Size() {
		return ErrOffsetExceedsFileSize
	}

	destFile, err := os.Create(toPath)
	if err != nil {
		return ErrOSError
	}
	defer destFile.Close()

	sourceFile.Seek(offset, io.SeekStart)
	bytesToCopy := fileInfo.Size() - offset
	if offset+limit < fileInfo.Size() {
		bytesToCopy = limit
	}

	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()
	for i := int64(0); i < offset; i++ {
		n, err := io.CopyN(destFile, sourceFile, 1)
		if err != nil {
			return ErrOSError
		}
		bar.Add64(n)
	}

	return nil
}
