package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require" //nolint:depguard,nolintlint
)

func TestRunCmd(t *testing.T) {
	t.Run("erroneous cases", func(t *testing.T) {
		erroneousCases(t)
	})

	t.Run("valid cases", func(t *testing.T) {
		validRunCmdCases(t)
	})
}

func erroneousCases(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name string
		cmd  []string
		env  Environment
		want int
	}{
		{
			name: "empty command",
			cmd:  []string{},
			env:  Environment{},
			want: ExitCodeFailure,
		},
		{
			name: "non-existing command",
			cmd:  []string{"nonexistentcommand"},
			env:  Environment{},
			want: ExitCodeFailure,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			result := RunCmd(tC.cmd, tC.env)
			require.Equal(t, tC.want, result, "unexpected return code")
		})
	}
}

func validRunCmdCases(t *testing.T) {
	t.Helper()

	// Test preparations.
	err := os.Setenv("EXISTING_VAR", "existing_value")
	if err != nil {
		require.Fail(t, "unable to set environment variable: "+err.Error())
	}
	defer os.Unsetenv("EXISTING_VAR")

	testCases := []struct {
		name       string
		cmd        []string
		env        Environment
		wantCode   int
		wantOutput string
	}{
		{
			name:       "simple command without args",
			cmd:        []string{"echo"},
			env:        Environment{},
			wantCode:   ExitCodeSuccess,
			wantOutput: "\n",
		},
		{
			name:       "command with args",
			cmd:        []string{"sh", "-c", "echo some text"},
			env:        Environment{},
			wantCode:   ExitCodeSuccess,
			wantOutput: "some text\n",
		},
		{
			name:       "add environment variable",
			cmd:        []string{"sh", "-c", "echo $TEST_VAR"},
			env:        Environment{"TEST_VAR": {Value: "test_value", NeedRemove: false}},
			wantCode:   ExitCodeSuccess,
			wantOutput: "test_value\n",
		},
		{
			name:       "change existing environment variable",
			cmd:        []string{"sh", "-c", "echo $EXISTING_VAR"},
			env:        Environment{"EXISTING_VAR": {Value: "some_new_value", NeedRemove: false}},
			wantCode:   ExitCodeSuccess,
			wantOutput: "some_new_value\n",
		},
		{
			name:       "remove existing environment variable",
			cmd:        []string{"sh", "-c", "echo $EXISTING_VAR"},
			env:        Environment{"EXISTING_VAR": {Value: "", NeedRemove: true}},
			wantCode:   ExitCodeSuccess,
			wantOutput: "\n",
		},
		{
			name:       "preserve existing environment",
			cmd:        []string{"sh", "-c", "echo $EXISTING_VAR"},
			env:        Environment{},
			wantCode:   ExitCodeSuccess,
			wantOutput: "existing_value\n",
		},
		{
			name:       "remove non-existing environment variable",
			cmd:        []string{"sh", "-c", "echo $NON_EXISTENT_VAR"},
			env:        Environment{"NON_EXISTENT_VAR": {Value: "", NeedRemove: true}},
			wantCode:   ExitCodeSuccess,
			wantOutput: "\n",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				require.Fail(t, "unable to create pipe: "+err.Error())
			}
			originalStdout := os.Stdout
			os.Stdout = w
			defer func() {
				os.Stdout = originalStdout
				w.Close()
			}()

			resCode := RunCmd(tC.cmd, tC.env)
			w.Close()

			var output strings.Builder
			_, err = io.Copy(&output, r)
			if err != nil {
				require.Fail(t, "unable to read output: "+err.Error())
			}

			// Error code check.
			require.Equal(t, tC.wantCode, resCode, "unexpected return code")

			// Output check.
			if tC.wantOutput != "" {
				require.Contains(t, output.String(), tC.wantOutput, "unexpected command output: "+output.String())
			}
		})
	}
}
