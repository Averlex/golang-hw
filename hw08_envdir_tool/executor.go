package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

//nolint:revive,nolintlint
const (
	ExitCodeSuccess = 0
	ExitCodeFailure = 1
)

func parseEnv(src []string) (env map[string]string) {
	env = make(map[string]string)

	for _, s := range src {
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			continue
		}
		env[kv[0]] = kv[1]
	}

	return
}

func mapToSlice(cmd map[string]string) (res []string) {
	res = make([]string, 0, len(cmd))
	for k, v := range cmd {
		res = append(res, k+"="+v)
	}
	return
}

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return ExitCodeFailure
	}

	var command *exec.Cmd
	//nolint:gosec
	if len(cmd) == 1 {
		command = exec.Command(cmd[0])
	} else {
		command = exec.Command(cmd[0], cmd[1:]...)
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	command.Env = make([]string, 0, len(env))
	newEnv := parseEnv(os.Environ())
	for k, v := range env {
		if v.NeedRemove {
			// Env variable is not present in the current environment.
			if _, ok := newEnv[k]; !ok {
				continue
			}
			delete(newEnv, k)
			continue
		}

		newEnv[k] = v.Value
	}

	command.Env = mapToSlice(newEnv)

	if err := command.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		return ExitCodeFailure
	}

	return ExitCodeSuccess
}
