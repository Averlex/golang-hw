// Package main contains entrypoint for the calendar service.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/app/calendar"         //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config"                   //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/logger"                   //nolint:depguard
	internalhttp "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/http" //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage"                  //nolint:depguard
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"          //nolint:depguard
)

const (
	exitCodeSuccess = 0
	exitCodeError   = 1
)

var defaultConfigFile = "../../configs/config.toml"

func main() {
	ctx := context.Background()
	// Creating temporary logger so errors are not lost.
	logg, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("create temporary logger: %s\n", err.Error())
		os.Exit(exitCodeError)
	}
	logg = logg.With("service", "calendar")

	// Loading configuration from file and env.
	loader := config.NewViperLoader("calendar", "Calendar service", "Calendar service for managing events and reminders",
		defaultConfigFile, "CALENDAR")
	cfg, err := loader.Load(printVersion, os.Stdout)
	if err != nil {
		if errors.Is(err, projectErrors.ErrShouldStop) {
			os.Exit(exitCodeSuccess)
		}
		logg.Fatal(ctx, "load config", "err", err)
	}

	// Wrapping logs with service name.
	logg = logg.With("service", "calendar")
	logg.Info(ctx, "config loaded successfully")

	// Initializing service logger.
	logCfg, err := cfg.GetSubConfig("logger")
	if err != nil {
		logg.Fatal(ctx, "get logger config", "err", err)
	}
	logg, err = logger.NewLogger(logger.WithConfig(logCfg))
	if err != nil {
		logg.Fatal(ctx, "create logger", "err", err, "config", logCfg)
	}
	logg.Info(ctx, "logger created successfully")

	// Initializing the storage.
	storageCfg, err := cfg.GetSubConfig("storage")
	if err != nil {
		logg.Fatal(ctx, "get storage config", "err", err)
	}
	storage, err := storage.NewStorage(storageCfg)
	if err != nil {
		logg.Fatal(ctx, "create storage", "err", err)
	}
	logg.Info(ctx, "storage created successfully")

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	defer storage.Close(ctx)

	// Initializing the app.
	appCfg, err := cfg.GetSubConfig("app")
	if err != nil {
		logg.Fatal(ctx, "get app config", "err", err)
	}
	calendar, err := app.NewApp(logg, storage, appCfg)
	if err != nil {
		logg.Fatal(ctx, "create app", "err", err)
	}
	logg.Info(ctx, "app created successfully")

	server := internalhttp.NewServer(logg, calendar)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error(ctx, "stop http server: "+err.Error())
		}
	}()

	logg.Info(ctx, "calendar is running...")

	if err := server.Start(ctx); err != nil {
		cancel()
		logg.Fatal(ctx, "start http server", "err", err)
	}
}
