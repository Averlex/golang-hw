package main

import (
	"encoding/json"
	"fmt"
	"io"
)

var (
	release   = "UNKNOWN"
	buildDate = "UNKNOWN"
	gitHash   = "UNKNOWN"
)

func printVersion(w io.Writer) error {
	if err := json.NewEncoder(w).Encode(struct {
		Release   string
		BuildDate string
		GitHash   string
	}{
		Release:   release,
		BuildDate: buildDate,
		GitHash:   gitHash,
	}); err != nil {
		return fmt.Errorf("error while decode version info: %w", err)
	}
	return nil
}
