package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3" //nolint:depguard
)

//nolint:revive,nolintlint
var (
	ErrUnsupportedFile       = errors.New("unsupported file")
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
	res := isSystemPath(path) || path == ""
	fileInfo, err := os.Stat(path)
	if err == nil {
		return res || fileInfo.IsDir()
	}
	return res
}

// Copy copies a file from a source path to a destination path with optional offset and limit.
// It returns an error if the source file is unsupported, the source or destination path is invalid,
// the offset exceeds the file size, or if an OS error occurs during the process.
// The function also creates the necessary directory structure for the destination path if it doesn't exist.
// Progress is shown using a progress bar.
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

	offset = max(0, offset)
	limit = max(0, limit)

	src, _ := os.Open(fromPath)
	defer src.Close()

	// Offset validation.
	fileInfo, _ := os.Stat(fromPath)
	if offset >= fileInfo.Size() {
		return ErrOffsetExceedsFileSize
	}

	// Creating directory tree if it doesn't exist.
	dir := filepath.Dir(toPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return ErrOSError
	}

	// Creating destination file.
	dst, err := os.Create(toPath)
	if err != nil {
		return ErrOSError
	}
	defer dst.Close()

	// Applying offset and calculating real limit.
	src.Seek(offset, io.SeekStart) // The possibility of using Seek() on src was checked above.
	bytesToCopy := fileInfo.Size() - offset
	if limit < fileInfo.Size()-offset && limit > 0 {
		bytesToCopy = limit
	}

	// Copying.
	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()
	reader := bar.NewProxyReader(src)
	_, err = io.CopyN(dst, reader, bytesToCopy)
	if err != nil && !errors.Is(err, io.EOF) {
		return ErrOSError
	}

	return nil
}
