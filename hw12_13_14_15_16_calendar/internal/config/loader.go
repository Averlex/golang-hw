package config

import (
	"fmt"
	"io"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
)

// Loader is a viper config loader.
type Loader struct {
	name, short, long string // Root command attributes.
	configPath        string
	envPrefix         string
}

// NewLoader returns a new viper loader.
//
// name is the short name of the service, as well as the name of the root command.
//
// short and long are the short and long descriptions of the service for the root command.
//
// configPath is the path to the configuration file.
//
// envPrefix is the prefix for environment variables.
func NewLoader(name, short, long, configPath, envPrefix string) *Loader {
	return &Loader{
		configPath: configPath,
		envPrefix:  envPrefix,
		name:       name,
		short:      short,
		long:       long,
	}
}

// Load loads configuration from a file and environment variables.
// It checks the flags in the process and executes the root command, depending on the flag.
// If no additional flags are set, it will load the configuration from the path provided.
//
// If -h (--help) or -v (--version) flags are set, it will return nil, ErrShouldStop
// as a signal to stop the execution.
func (l *Loader) Load(printVersion func(io.Writer) error, writer io.Writer) (ServiceConfig, error) {
	cfg, err := NewServiceConfig(l.name)
	if err != nil {
		return nil, fmt.Errorf("declare config: %w", err)
	}

	cmd, err := l.buildRootCommand(l.name, l.short, l.long, cfg, printVersion, writer)
	if err != nil {
		return nil, fmt.Errorf("build root command: %w", err)
	}

	if err := cmd.Execute(); err != nil {
		return nil, fmt.Errorf("execute root command: %w", err)
	}

	// Check if the help or version flag is set. If so, stop execution.
	if cmd.Flags().Changed("help") || cmd.Flags().Changed("version") {
		return nil, errors.ErrShouldStop
	}

	return cfg, nil
}
