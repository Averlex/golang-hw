//nolint:revive,nolintlint
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/app"                      //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config"                   //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/logger"                   //nolint:depguard
	internalhttp "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/http" //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage"                  //nolint:depguard
)

const (
	exitCodeSuccess = 0
	exitCodeError   = 1
)

var defaultConfigFile = "../../configs/config.toml"

func main() {
	// Creating temporary logger so errors are not lost.
	logg, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("failed to create logger: %s\n", err.Error())
		os.Exit(exitCodeError)
	}

	// Loading configuration from file and env.
	loader := config.NewViperLoader("calendar", "Calendar service", "Calendar service for managing events and reminders",
		defaultConfigFile, "CALENDAR")
	cfg, err := loader.Load(printVersion, os.Stdout)
	if err != nil {
		logg.Fatal("failed to load config", "err", err)
	}
	logg.Debug("config loaded successfully")

	// Initializing service logger.
	logCfg, err := cfg.GetSubConfig("logger")
	if err != nil {
		logg.Fatal("failed to get logger config", "err", err)
	}
	logg, err = logger.NewLogger(logger.WithConfig(logCfg))
	if err != nil {
		logg.Fatal("failed to create logger", "err", err, "config", logCfg)
	}
	logg.Debug("logger created successfully")

	// Initializing the storage.
	storageCfg, err := cfg.GetSubConfig("storage")
	if err != nil {
		logg.Fatal("failed to get storage config", "err", err)
	}
	storage, err := storage.NewStorage(storageCfg)
	if err != nil {
		logg.Fatal("failed to create storage", "config", storageCfg, "err", err)
	}

	// App and .....
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
		logg.Fatal("failed to start http server", "err", err)
	}
}
