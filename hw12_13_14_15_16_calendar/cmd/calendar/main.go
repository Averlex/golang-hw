// Package main contains entrypoint for the calendar service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	app "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/app/calendar"                  //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config"                            //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/logger"                            //nolint:depguard
	internalgrpc "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/server/calendar/grpc" //nolint:depguard
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
	if err := run(); err != nil {
		os.Exit(exitCodeError)
	}
	os.Exit(exitCodeSuccess)
}

func run() error {
	ctx := context.Background()
	// Creating temporary logger so errors will not be lost.
	logg, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("create temporary logger: %s\n", err.Error())
		return err
	}
	logg = logg.With(slog.String("service", "calendar"))

	// Loading configuration from file and env.
	cfg, err := loadConfig(ctx, logg)
	if err != nil {
		return err
	}

	// Initializing service logger.
	logg, err = initializeLogger(ctx, logg, cfg)
	if err != nil {
		return err
	}

	// Initializing the storage.
	storage, err := initializeStorage(ctx, logg, cfg)
	if err != nil {
		return err
	}
	defer storage.Close(ctx)

	// Initializing signal handler.
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Initializing storage connection.
	if err := storage.Connect(ctx); err != nil {
		logg.Error(ctx, "connect storage", slog.Any("err", err))
		return err
	}
	logg.Info(ctx, "storage connected established")

	// Initializing the app.
	calendar, err := initializeApp(ctx, logg, cfg, storage)
	if err != nil {
		return err
	}

	// Starting servers.
	return startServers(ctx, cancel, logg, cfg, calendar)
}

func loadConfig(ctx context.Context, logg *logger.Logger) (config.ServiceConfig, error) {
	loader := config.NewLoader("calendar", "Calendar service", "Calendar service for managing events and reminders",
		defaultConfigFile, "CALENDAR")
	cfg, err := loader.Load(printVersion, os.Stdout)
	if err != nil {
		if errors.Is(err, projectErrors.ErrShouldStop) {
			return nil, nil
		}
		logg.Error(ctx, "load config", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "config loaded successfully")
	return cfg, nil
}

func initializeLogger(ctx context.Context, logg *logger.Logger, cfg config.ServiceConfig) (*logger.Logger, error) {
	logCfg, err := cfg.GetSubConfig("logger")
	if err != nil {
		logg.Error(ctx, "get logger config", slog.Any("err", err))
		return nil, err
	}
	newLogg, err := logger.NewLogger(logger.WithConfig(logCfg))
	if err != nil {
		logg.Error(ctx, "create logger", slog.Any("err", err))
		return nil, err
	}
	newLogg = newLogg.With(slog.String("service", "calendar"))
	newLogg.Info(ctx, "logger created successfully")
	return newLogg, nil
}

func initializeStorage(ctx context.Context, logg *logger.Logger, cfg config.ServiceConfig) (storage.Storage, error) {
	storageCfg, err := cfg.GetSubConfig("storage")
	if err != nil {
		logg.Error(ctx, "get storage config", slog.Any("err", err))
		return nil, err
	}
	storage, err := storage.NewStorage(storageCfg)
	if err != nil {
		logg.Error(ctx, "create storage", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "storage created successfully")
	return storage, nil
}

func initializeApp(
	ctx context.Context,
	logg *logger.Logger,
	cfg config.ServiceConfig,
	storage storage.Storage,
) (*app.App, error) {
	appCfg, err := cfg.GetSubConfig("app")
	if err != nil {
		logg.Error(ctx, "get app config", slog.Any("err", err))
		return nil, err
	}
	calendar, err := app.NewApp(logg.With(slog.String("layer", "APP")), storage, appCfg)
	if err != nil {
		logg.Error(ctx, "create app", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "app created successfully")
	return calendar, nil
}

func startServers(
	ctx context.Context,
	cancel context.CancelFunc,
	logg *logger.Logger,
	cfg config.ServiceConfig,
	calendar *app.App,
) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2) // Buffer for Start errors of both servers.

	grpcServer, err := initializeGRPCServer(ctx, logg.With(slog.String("layer", "gRPC server")), cfg, calendar)
	if err != nil {
		return err
	}

	httpServer, err := initializeHTTPServer(ctx, logg.With(slog.String("layer", "HTTP server")), cfg, calendar)
	if err != nil {
		return err
	}

	// Starting gRPC server.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Start(ctx); err != nil {
			logg.Error(ctx, "start gRPC server", slog.Any("err", err))
			errChan <- err
		}
	}()

	// Starting http server.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.Start(ctx); err != nil {
			logg.Error(ctx, "start HTTP server", slog.Any("err", err))
			errChan <- err
		}
	}()

	logg.Info(ctx, "calendar is running...")

	// Wait for the first error or context cancellation.
	var serverErr error
	select {
	// We assume that the only possible result of server start is error or blocking.
	case serverErr = <-errChan:
		cancel()
		logg.Error(ctx, "server error occurred, shutting down calendar...")
	case <-ctx.Done():
		logg.Info(ctx, "interruption received, shutting down calendar...")
	}

	// Stop both servers explicitly.
	if err := grpcServer.Stop(ctx); err != nil {
		logg.Error(ctx, "stop gRPC server", slog.Any("err", err))
	} else {
		logg.Info(ctx, "gRPC server stopped successfully")
	}
	if err := httpServer.Stop(ctx); err != nil {
		logg.Error(ctx, "stop HTTP server", slog.Any("err", err))
	} else {
		logg.Info(ctx, "HTTP server stopped successfully")
	}

	wg.Wait()
	close(errChan)
	return serverErr
}

func initializeGRPCServer(
	ctx context.Context,
	logg *logger.Logger,
	cfg config.ServiceConfig,
	calendar *app.App,
) (*internalgrpc.Server, error) {
	grpcCfg, err := cfg.GetSubConfig("grpc")
	if err != nil {
		logg.Error(ctx, "get gRPC server config", slog.Any("err", err))
		return nil, err
	}
	grpcServer, err := internalgrpc.NewServer(logg, calendar, grpcCfg)
	if err != nil {
		logg.Error(ctx, "create gRPC server", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "gRPC server created successfully")
	return grpcServer, nil
}

func initializeHTTPServer(
	ctx context.Context,
	logg *logger.Logger,
	cfg config.ServiceConfig,
	calendar *app.App,
) (*internalhttp.Server, error) {
	httpCfg, err := cfg.GetSubConfig("http")
	if err != nil {
		logg.Error(ctx, "get HTTP server config", slog.Any("err", err))
		return nil, err
	}
	grpcCfg, err := cfg.GetSubConfig("grpc")
	if err != nil {
		logg.Error(ctx, "get gRPC server config for HTTP server", slog.Any("err", err))
		return nil, err
	}
	httpServer, err := internalhttp.NewServer(logg, calendar, httpCfg, grpcCfg)
	if err != nil {
		logg.Error(ctx, "create HTTP server", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "HTTP server created successfully")
	return httpServer, nil
}
