// Package main contains two functions: ReadDir() and RunCmd().
package main

import "os"

func main() {
	args := os.Args
	if len(args) == 1 || len(args) == 2 {
		os.Exit(ExitCodeFailure)
	}
	env, err := ReadDir(args[1])
	if err != nil {
		os.Exit(ExitCodeFailure)
	}
	os.Exit(RunCmd(args[2:], env))
}
