package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//nolint:revive,nolintlint
type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

//nolint:revive,nolintlint
var (
	ErrIncorrectPath   = errors.New("incorrect path")
	ErrFSError         = errors.New("file system error")
	ErrUnsupportedFile = errors.New("unsupported file")
)

var unsupportedPathPrefixes = []string{
	"/dev",
	"/proc",
	"/sys",
	"/run",
	"/var/run",
	"/var/lock",
}

func isIncorrectPath(path string) bool {
	fileInfo, err := os.Lstat(path)
	// File not exists.
	if err != nil {
		return true
	}

	// Not a directory or a symlink.
	if !fileInfo.IsDir() || (fileInfo.Mode()&os.ModeSymlink != 0) {
		return true
	}

	return false
}

func isSystemPath(path string) bool {
	for _, prefix := range unsupportedPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
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

func readFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}

func removeExt(fname string) string {
	return strings.TrimSuffix(fname, filepath.Ext(fname))
}

func containsEqualSign(s string) bool { return strings.Contains(s, "=") }

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	if isIncorrectPath(dir) || isSystemPath(dir) {
		return nil, ErrIncorrectPath
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, ErrFSError
	}

	env := make(Environment)

	for _, entry := range entries {
		path := dir + "/" + entry.Name()
		if isUnsupportedFile(path) {
			return nil, ErrUnsupportedFile
		}

		line, err := readFile(path)
		if err != nil {
			return nil, ErrFSError
		}

		if containsEqualSign(line) {
			return nil, ErrIncorrectPath
		}

		line = strings.TrimRight(line, " \t")
		line = strings.ReplaceAll(line, "\x00", "\n")

		needRemove := line == ""
		env[removeExt(entry.Name())] = EnvValue{Value: line, NeedRemove: needRemove}
	}

	return env, nil
}
