package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require" //nolint:depguard,nolintlint
)

func TestCopy(t *testing.T) {
	testDirPath := "./tmp"
	if _, err := os.Stat(testDirPath); err == nil {
		os.RemoveAll(testDirPath)
	}
	os.Mkdir(testDirPath, os.ModePerm)
	defer os.RemoveAll(testDirPath)

	incorrectFile(t, testDirPath)
	incorrectPath(t, testDirPath)
}

func incorrectFile(t *testing.T, testDirPath string) {
	t.Helper()

	t.Run("incorrect source file", func(t *testing.T) {
		want := ErrUnsupportedFile
		output := testDirPath + "/output.txt"
		testCases := []struct {
			name  string
			fname string
		}{
			{"does not exist", testDirPath + "/some_nonexistent_file.txt"},
			{"not regular", testDirPath + "/folder"},
			{"symlink", testDirPath + "/symlink.txt"},
			{"undefined file size", testDirPath + "/undefined_size.txt"},
		}

		// Test preparations.
		os.Mkdir(testDirPath+"/folder", os.ModePerm)
		if err := os.Symlink("testdata/input.txt", testCases[2].fname); err != nil {
			require.Fail(t, "unable to create symlink: "+err.Error())
		}
		if err := os.WriteFile(testCases[3].fname, []byte("42 is the anwser"), os.ModePerm); err != nil {
			require.Fail(t, "unable to create a file with undefined size: "+err.Error())
		}
		if err := os.Chmod(testCases[3].fname, os.ModeDevice); err != nil {
			require.Fail(t, "unable to change file permissions for a file with undefined size: "+err.Error())
		}

		for _, tC := range testCases {
			t.Run(tC.name, func(t *testing.T) {
				res := Copy(tC.fname, output, 0, 0)
				if res == nil {
					require.Fail(t, "expected error, got nil")
				}
				require.ErrorIs(t, res, want)
			})
		}
	})
}

func incorrectPath(t *testing.T, testDirPath string) {
	t.Helper()

	want := ErrUnsupportedPath
	validInput := "./testdata/input.txt"
	validOutput := "./testdata/output.txt"
	testCases := []struct {
		name   string
		source string
	}{
		{"proc", "/proc/self/cmdline"},
		{"sys", "/sys/devices/system/cpu/online"},
		{"run", "/run/utmp"},
	}

	// Test preparations.
	folderPath := testDirPath + "/some_folder"
	err := os.Mkdir(folderPath, os.ModePerm)
	if err != nil {
		require.Fail(t, "unable to create a folder: "+err.Error())
	}

	t.Run("system source path", func(t *testing.T) {
		for _, tC := range testCases {
			t.Run(tC.name, func(t *testing.T) {
				if tC.name == "proc" {
					fmt.Println("proc")
				}
				res := Copy(tC.source, validOutput, 0, 0)
				if res == nil {
					require.Fail(t, "expected error, got nil")
				}
				require.ErrorIs(t, res, want)
			})
		}
	})

	t.Run("invalid destination path", func(t *testing.T) {
		for _, tC := range testCases {
			t.Run(tC.name, func(t *testing.T) {
				res := Copy(validInput, tC.source, 0, 0)
				if res == nil {
					require.Fail(t, "expected error, got nil")
				}
				require.ErrorIs(t, res, want)
			})
		}
		t.Run("dev/null", func(t *testing.T) {
			res := Copy(validInput, folderPath, 0, 0)
			if res == nil {
				require.Fail(t, "expected error, got nil")
			}
			require.ErrorIs(t, res, want)
		})
		t.Run("toPath is a folder", func(t *testing.T) {
			res := Copy(validInput, folderPath, 0, 0)
			if res == nil {
				require.Fail(t, "expected error, got nil")
			}
			require.ErrorIs(t, res, want)
		})
	})
}
