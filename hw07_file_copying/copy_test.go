package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require" //nolint:depguard,nolintlint
)

const defaultInputFile = "./testdata/input.txt"

func TestCopy(t *testing.T) {
	testDirPath := "./tmp"
	if _, err := os.Stat(testDirPath); err == nil {
		os.RemoveAll(testDirPath)
	}
	os.Mkdir(testDirPath, os.ModePerm)
	defer os.RemoveAll(testDirPath)

	incorrectFile(t, testDirPath)
	incorrectPath(t, testDirPath)
	defaultTestdata(t, testDirPath)
	offsetLimit(t, testDirPath)

	t.Run("nested folder creation", func(t *testing.T) {
		source := defaultInputFile
		templateOutput := "./testdata/out_offset0_limit0.txt"
		output := testDirPath + "/folder/another_folder/one_more/output.txt"
		res := Copy(source, output, 0, 0)
		if res != nil {
			require.Fail(t, "expected nil, got error: "+res.Error())
		}
		compare(t, output, templateOutput)
	})
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
	validInput := defaultInputFile
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
				res := Copy(tC.source, validOutput, 0, 0)
				if res == nil {
					require.Fail(t, "expected error, got nil")
				}
				require.ErrorIs(t, res, want)
			})
		}
		t.Run("empty source path", func(t *testing.T) {
			res := Copy("", validOutput, 0, 0)
			if res == nil {
				require.Fail(t, "expected error, got nil")
			}
			require.ErrorIs(t, res, ErrUnsupportedFile)
		})
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
		t.Run("empty source path", func(t *testing.T) {
			res := Copy(validInput, "", 0, 0)
			if res == nil {
				require.Fail(t, "expected error, got nil")
			}
			require.ErrorIs(t, res, want)
		})
	})
}

func defaultTestdata(t *testing.T, testDirPath string) {
	t.Helper()

	t.Run("testdata directory default contents", func(t *testing.T) {
		source := defaultInputFile
		testCases := []struct {
			name           string
			templateOutput string
			output         string
			offset         int64
			limit          int64
		}{
			{
				"offset 0, limit 0", "./testdata/out_offset0_limit0.txt",
				testDirPath + "/out_offset0_limit0.txt", 0, 0,
			},
			{
				"offset 0, limit 10", "./testdata/out_offset0_limit10.txt",
				testDirPath + "/out_offset0_limit10.txt", 0, 10,
			},
			{
				"offset 0, limit 1000", "./testdata/out_offset0_limit1000.txt",
				testDirPath + "/out_offset0_limit1000.txt", 0, 1000,
			},
			{
				"offset 0, limit 10000", "./testdata/out_offset0_limit10000.txt",
				testDirPath + "/out_offset0_limit10000.txt", 0, 10000,
			},
			{
				"offset 100, limit 1000", "./testdata/out_offset100_limit1000.txt",
				testDirPath + "/out_offset100_limit1000.txt", 100, 1000,
			},
			{
				"offset 6000, limit 1000", "./testdata/out_offset6000_limit1000.txt",
				testDirPath + "/out_offset6000_limit1000.txt", 6000, 1000,
			},
		}
		for _, tC := range testCases {
			t.Run(tC.name, func(t *testing.T) {
				Copy(source, tC.output, tC.offset, tC.limit)
				compare(t, tC.output, tC.templateOutput)
			})
		}
	})
}

func offsetLimit(t *testing.T, testDirPath string) {
	t.Helper()

	t.Run("offset and limit", func(t *testing.T) {
		source := defaultInputFile
		output := testDirPath + "/output.txt"

		testCases := []struct {
			name   string
			offset int64
			limit  int64
		}{
			{"negative offset", -1, 0},
			{"negative limit", 0, -1},
		}

		for _, tC := range testCases {
			t.Run(tC.name, func(t *testing.T) {
				res := Copy(source, output, tC.offset, tC.limit)
				if res != nil {
					require.Fail(t, "expected nil, got error: "+res.Error())
				}
				compare(t, output, "./testdata/out_offset0_limit0.txt")
			})
		}

		t.Run("offset greater than file size", func(t *testing.T) {
			res := Copy(source, output, 1000000, 0)
			if res == nil {
				require.Fail(t, "expected error, got nil")
			}
			require.ErrorIs(t, res, ErrOffsetExceedsFileSize)
		})
	})
}

func compare(t *testing.T, dst, tmplDst string) {
	t.Helper()

	output, err := os.ReadFile(dst)
	if err != nil {
		require.Fail(t, "unable to read output file: "+err.Error())
	}
	templateOutput, err := os.ReadFile(tmplDst)
	if err != nil {
		require.Fail(t, "unable to read template output file: "+err.Error())
	}
	require.Equal(t, output, templateOutput)
}
