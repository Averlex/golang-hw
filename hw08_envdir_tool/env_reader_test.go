package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const defaultDirPath = "./testdata"

func TestReadDir(t *testing.T) {
	testDirPath := "./tmp"
	if _, err := os.Stat(testDirPath); err == nil {
		os.RemoveAll(testDirPath)
	}
	os.Mkdir(testDirPath, os.ModePerm)
	defer os.RemoveAll(testDirPath)

	t.Run("erroneous cases", func(t *testing.T) {
		incorrectFolderContents(t, testDirPath)
	})

	// file with extension
	// file with trailing spaces/tabs
	// file with trailing spaces/tabs in the beginning
	// file with \x00
	// empty file
	// file with multiple lines
	// testdata case

	// t.Run("valid cases", func(t *testing.T) {
	// })
}

func incorrectFolderContents(t *testing.T, testDirPath string) {
	t.Helper()

	testCases := []struct {
		name string
		path string
		want error
	}{
		{"not a folder", testDirPath + "/some_nonexistent_file.txt", ErrIncorrectPath},
		{"symlink", testDirPath + "/symlink", ErrIncorrectPath},
		{"contains folders", testDirPath + "/folders", ErrUnsupportedFile},
		{"contains devices", testDirPath + "/devices", ErrUnsupportedFile},
		{"empty folder", testDirPath + "/empty_folder", ErrFSError},
		{"file name contains '=' character", testDirPath + "/incorrect_file_name", ErrUnsupportedFile},
	}

	// Test preparations.
	if err := os.WriteFile(testCases[0].path, []byte("42"), os.ModePerm); err != nil {
		require.Fail(t, "unable to create a file: "+err.Error())
	}

	// Skipping file and symlink cases.
	for i := 2; i < len(testCases); i++ {
		if err := os.Mkdir(testCases[i].path, os.ModePerm); err != nil {
			require.Fail(t, "unable to create a folder: "+err.Error())
		}
	}

	if err := os.Symlink(defaultDirPath, testCases[1].path); err != nil {
		require.Fail(t, "unable to create symlink: "+err.Error())
	}
	if err := os.Mkdir(testCases[2].path+"/additional_folder", os.ModePerm); err != nil {
		require.Fail(t, "unable to create a folder: "+err.Error())
	}
	if err := os.WriteFile(testCases[3].path+"/empty_device.txt", []byte(""), os.ModePerm); err != nil {
		require.Fail(t, "unable to create a file: "+err.Error())
	}
	if err := os.Chmod(testCases[3].path+"/empty_device.txt", os.ModeDevice); err != nil {
		require.Fail(t, "unable to change file permissions to device: "+err.Error())
	}
	if err := os.WriteFile(testCases[5].path+"/incorrect_file_name=.txt", []byte(""), os.ModePerm); err != nil {
		require.Fail(t, "unable to create a file: "+err.Error())
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			env, err := ReadDir(tC.path)
			require.ErrorIs(t, err, tC.want, "unexpected error received: "+err.Error())
			require.Nil(t, env, fmt.Sprintf("nil env expected, but it isn't: %T, %v", env, env))
		})
	}
}
