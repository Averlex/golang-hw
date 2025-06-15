package config

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra" //nolint:depguard,nolintlint
	"github.com/spf13/viper" //nolint:depguard,nolintlint
)

// buildRootCommand builds root command.
//
// Method declares flags and binds them to actions. It also enables env variables.
// If any of the env variables is set, it will overrride the config file.
//
// The result of the execution - is a built up command, ready to execute.
// All reading/setting is perfromed on the pre-run stage.
func (l *Loader) buildRootCommand(name, short, long string,
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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.BindPFlag("config", rootCmd.Flags().Lookup("config")); err != nil {
		return nil, fmt.Errorf("bind config flag: %w", err)
	}

	if err := viper.BindPFlag("version", rootCmd.Flags().Lookup("version")); err != nil {
		return nil, fmt.Errorf("bind version flag: %w", err)
	}
	// viper.Debug()

	rootCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		// Processing -v flag preemptively.
		if versionFlag := viper.GetBool("version"); versionFlag {
			if err := printVersion(writer); err != nil {
				return fmt.Errorf("print version: %w", err)
			}
			return nil
		}

		// Setting the config.
		configPath := viper.GetString("config")
		viper.SetConfigFile(configPath)
		viper.ReadInConfig()

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("read main config at %s: %w", configPath, err)
		}
		if err := viper.Unmarshal(cfg); err != nil {
			return fmt.Errorf("unmarshal main config: %w", err)
		}

		// viper.Debug()
		return nil
	}

	return rootCmd, nil
}
