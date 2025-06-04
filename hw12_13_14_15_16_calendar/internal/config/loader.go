package config

import (
	"fmt"
	"io"

	"github.com/spf13/cobra" //nolint:depguard,nolintlint
	"github.com/spf13/viper" //nolint:depguard,nolintlint
)

// Loader is a configuration loader.
type Loader interface {
	Load() (map[string]any, error)
}

// ViperLoader is a viper config loader.
type ViperLoader struct {
	name, short, long string // Root command attributes.
	configPath        string
	envPrefix         string
}

// NewViperLoader returns a new viper loader.
//
// name is the short name of the service, as well as the name of the root command.
//
// short and long are the short and long descriptions of the service for the root command.
//
// configPath is the path to the configuration file.
//
// envPrefix is the prefix for environment variables.
func NewViperLoader(name, short, long, configPath, envPrefix string) *ViperLoader {
	return &ViperLoader{
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
func (l *ViperLoader) Load(printVersion func(io.Writer) error, writer io.Writer) (ServiceConfig, error) {
	cfg, err := NewServiceConfig(l.name)
	if err != nil {
		return nil, fmt.Errorf("failed to declare config: %w", err)
	}

	cmd, err := l.buildRootCommand(l.name, l.short, l.long, cfg, printVersion, writer)
	if err != nil {
		return nil, fmt.Errorf("failed to build root command: %w", err)
	}

	if err := cmd.Execute(); err != nil {
		return nil, fmt.Errorf("failed to execute root command: %w", err)
	}

	return cfg, nil
}

func (l *ViperLoader) buildRootCommand(name, short, long string,
	cfg ServiceConfig,
	printVersion func(io.Writer) error, writer io.Writer,
) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  long,
		Run: func(_ *cobra.Command, _ []string) {
			// Service logic will be handled in another place.
		},
	}

	rootCmd.Flags().StringP("config", "c", "", "Path to configuration file")
	rootCmd.Flags().BoolP("version", "v", false, "Show version info")

	viper.SetEnvPrefix(l.envPrefix)
	viper.AutomaticEnv()

	if err := viper.BindPFlag("config", rootCmd.Flags().Lookup("config")); err != nil {
		return nil, fmt.Errorf("failed to bind config flag: %w", err)
	}

	if err := viper.BindPFlag("version", rootCmd.Flags().Lookup("version")); err != nil {
		return nil, fmt.Errorf("failed to bind version flag: %w", err)
	}

	rootCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		// Processing -v flag preemptively.
		if versionFlag := viper.GetBool("version"); versionFlag {
			if err := printVersion(writer); err != nil {
				return fmt.Errorf("failed to print version: %w", err)
			}
			return nil
		}

		// Setting the config.
		configPath := viper.GetString("config")
		viper.SetConfigFile(configPath)

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read main config at %s: %w", configPath, err)
		}
		if err := viper.Unmarshal(cfg); err != nil {
			return fmt.Errorf("failed to unmarshal main config: %w", err)
		}
		return nil
	}

	return rootCmd, nil
}
