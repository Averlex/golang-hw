package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require" //nolint:depguard
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

	t.Run("valid cases", func(t *testing.T) {
		validCases(t, testDirPath)
	})
}

func incorrectFolderContents(t *testing.T, testDirPath string) {
	t.Helper()

	testCases := []struct {
		name string
		path string
		want error
	}{
		{"not a folder", testDirPath + "/not_a_folder.txt", ErrIncorrectPath},
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

func validCases(t *testing.T, testDirPath string) {
	t.Helper()

	path := testDirPath + "/valid_cases"
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		require.Fail(t, "unable to create a folder: "+err.Error())
	}

	testCases := []struct {
		name     string
		fname    string
		contents string
		want     string
	}{
		{"file with extension", "file_with_extension.txt", "", ""},
		{"file with trailing spaces/tabs", "file_with_trailing_spaces", "contents\t    ", "contents"},
		{
			"file with trailing spaces in the beginning",
			"file_with_trailing_spaces_in_the_beginning",
			" \t contents", " \t contents",
		},
		{"file with '\\x00'", "file_with_x00", "conte\x00n\x00ts\x00", "conte\nn\nts\n"},
		{"empty file", "empty_file", "", ""},
		{"file with multiple lines", "file_with_multiple_lines", "line1\nline2\nline3\n", "line1"},
	}

	for _, tC := range testCases {
		if err := os.WriteFile(path+"/"+tC.fname, []byte(tC.contents), os.ModePerm); err != nil {
			require.Fail(t, "unable to create a file: "+err.Error())
		}
	}

	env, err := ReadDir(path)
	require.NoError(t, err, "unexpected error received")
	require.NotNil(t, env, fmt.Sprintf("nil env expected, but it isn't: %T, %v", env, env))
	require.Len(t, env, len(testCases), "unexpected env length")

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			fname := tC.fname
			if tC.name == "file with extension" {
				fname = "file_with_extension"
			}
			require.Equal(t, tC.want, env[fname].Value, "unexpected value received")

			if tC.name == "empty file" {
				require.True(t, env[fname].NeedRemove, "unexpected value received")
			}
		})
	}
}
