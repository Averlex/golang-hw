//nolint:revive,nolintlint
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_calendar/internal/app"                          //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_calendar/internal/logger"                       //nolint:depguard
	internalhttp "github.com/Averlex/golang-hw/hw12_13_14_15_calendar/internal/server/http"     //nolint:depguard
	memorystorage "github.com/Averlex/golang-hw/hw12_13_14_15_calendar/internal/storage/memory" //nolint:depguard
	"github.com/spf13/cobra"                                                                    //nolint:depguard
	"github.com/spf13/viper"                                                                    //nolint:depguard
)

const (
	exitCodeSuccess   = 0
	exitCodeError     = 1
	defaultConfigFile = "/configs/config.toml"
)

func main() {
	var cfg *Config
	tmpLogger, err := logger.New("", "error", "prod", os.Stdout)
	if err != nil {
		log.Fatalf("failed to create temporarylogger: %v", err)
	}
	rootCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar service",
		Long:  "Calendar service for managing events and reminders",
		Run: func(_ *cobra.Command, _ []string) {
			if err := run(cfg); err != nil {
				tmpLogger.Error("failed to run service", "error", err)
				os.Exit(exitCodeError)
			}
		},
	}

	rootCmd.Flags().StringP("config", "c", defaultConfigFile, "Path to configuration file")
	rootCmd.Flags().BoolP("version", "v", false, "Show version info")

	if err := viper.BindPFlag("config", rootCmd.Flags().Lookup("config")); err != nil {
		tmpLogger.Error("failed to bind config flag", "error", err)
		os.Exit(exitCodeError)
	}

	rootCmd.PreRun = func(_ *cobra.Command, _ []string) {
		// Processing -v flag preemptively.
		if versionFlag := viper.GetBool("version"); versionFlag {
			if err := printVersion(os.Stdout); err != nil {
				tmpLogger.Error("failed to print version", "error", err)
				os.Exit(exitCodeError)
			}
			os.Exit(exitCodeSuccess)
		}

		// Setting the config.
		configPath := viper.GetString("config")
		if configPath == "" {
			configPath = defaultConfigFile
		}
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			tmpLogger.Error("failed to read main config", "error", err, "path", configPath)
			os.Exit(exitCodeError)
		}
		if err := viper.Unmarshal(cfg); err != nil {
			tmpLogger.Error("failed to unmarshal main config", "error", err)
			os.Exit(exitCodeError)
		}
		viper.AutomaticEnv()
	}

	if err := rootCmd.Execute(); err != nil {
		tmpLogger.Error("failed to execute root command", "error", err)
		os.Exit(exitCodeError)
	}
}

func run(cfg *Config) error {
	var w io.Writer
	switch cfg.App.LogStream {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		return fmt.Errorf("unknown log stream: %s", cfg.App.LogStream)
	}

	logg, err := logger.New(cfg.Logger.Level, cfg.Logger.Format, cfg.App.Env, w)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	logg.Debug("logger created successfully")

	// TODO: обернуть в фабрику стораджей
	storage := memorystorage.New() // TODO: choose depending on flags
	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to start http server: %w", err)
	}

	return nil
}
