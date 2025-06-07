// Package main contains entrypoint for the calendar service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	app "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/app/calendar"                  //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config"                            //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/logger"                            //nolint:depguard
	internalhttp "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/calendar/http" //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage"                           //nolint:depguard
	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors"                   //nolint:depguard
)

const (
	exitCodeSuccess = 0
	exitCodeError   = 1
)

var defaultConfigFile = "../../configs/calendar/config.toml"

func main() {
	ctx := context.Background()
	// Creating temporary logger so errors will not be lost.
	logg, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("create temporary logger: %s\n", err.Error())
		os.Exit(exitCodeError)
	}
	logg = logg.With(slog.String("service", "calendar"))

	// Loading configuration from file and env.
	loader := config.NewLoader("calendar", "Calendar service", "Calendar service for managing events and reminders",
		defaultConfigFile, "CALENDAR")
	cfg, err := loader.Load(printVersion, os.Stdout)
	if err != nil {
		if errors.Is(err, projectErrors.ErrShouldStop) {
			os.Exit(exitCodeSuccess)
		}
		logg.Fatal(ctx, "load config", slog.Any("err", err))
	}
	logg.Info(ctx, "config loaded successfully")

	// Initializing service logger.
	logCfg, err := cfg.GetSubConfig("logger")
	if err != nil {
		logg.Fatal(ctx, "get logger config", slog.Any("err", err))
	}
	logg, err = logger.NewLogger(logger.WithConfig(logCfg))
	if err != nil {
		logg.Fatal(ctx, "create logger", slog.Any("err", err))
	}
	logg = logg.With(slog.String("service", "calendar"))
	logg.Info(ctx, "logger created successfully")

	// Initializing the storage.
	storageCfg, err := cfg.GetSubConfig("storage")
	if err != nil {
		logg.Fatal(ctx, "get storage config", slog.Any("err", err))
	}
	user, pw := os.Getenv("CALENDAR_STORAGE_SQL_USER"), os.Getenv("CALENDAR_STORAGE_SQL_PASSWORD")
	storageCfg["sql"]["user"] = user
	storageCfg["sql"]["password"] = pw
	cfg.SetSecrets(user, pw)

	logg.Warn(ctx, "debug output", slog.String("DB_HOST", os.Getenv("CALENDAR_STORAGE_SQL_USER")), slog.String("DB_PW", os.Getenv("CALENDAR_STORAGE_SQL_PASSWORD")))
	logg.Warn(ctx, "debug output", slog.String("DB_HOST", storageCfg["host"].(string)), slog.String("DB_PW", storageCfg["password"].(string)))

	storage, err := storage.NewStorage(storageCfg)
	if err != nil {
		logg.Fatal(ctx, "create storage", slog.Any("err", err))
	}
	logg.Info(ctx, "storage created successfully")

	// Initializing signal handler.
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Initializing storage connection.
	err = storage.Connect(ctx)
	if err != nil {
		logg.Fatal(ctx, "connect storage", slog.Any("err", err))
	}
	logg.Info(ctx, "storage connected established")
	defer storage.Close(ctx)

	// Initializing the app.
	appCfg, err := cfg.GetSubConfig("app")
	if err != nil {
		logg.Fatal(ctx, "get app config", slog.Any("err", err))
	}
	calendar, err := app.NewApp(logg, storage, appCfg)
	if err != nil {
		logg.Fatal(ctx, "create app", slog.Any("err", err))
	}
	logg.Info(ctx, "app created successfully")

	// Starting http server.
	serverCfg, err := cfg.GetSubConfig("server")
	if err != nil {
		logg.Fatal(ctx, "get server config", slog.Any("err", err))
	}
	server, err := internalhttp.NewServer(logg, calendar, serverCfg)
	if err != nil {
		logg.Fatal(ctx, "create http server", slog.Any("err", err))
	}
	logg.Info(ctx, "server created successfully")

	go func() {
		<-ctx.Done()

		if err := server.Stop(ctx); err != nil {
			logg.Error(ctx, "stop http server", slog.Any("err", err))
		}
	}()

	logg.Info(ctx, "calendar is running...")

	if err := server.Start(ctx); err != nil {
		cancel()
		logg.Fatal(ctx, "start http server", slog.Any("err", err))
	}
}
