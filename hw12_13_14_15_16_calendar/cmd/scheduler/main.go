// Package main contains entrypoint for the calendar scheduler service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	schedulerConfig "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config/scheduler" //nolint:depguard
	schedulerPkg "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/scheduler"           //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage"                          //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/config"                                //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/logger"                                //nolint:depguard
	mq "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/rabbitmq"                           //nolint:depguard
)

const (
	exitCodeSuccess = 0
	exitCodeError   = 1
)

var defaultConfigFile = "../../configs/scheduler/config.toml"

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
	logg = logg.With(slog.String("service", "scheduler"))

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
	logg.Info(ctx, "storage connection established")

	// Initializing message queue.
	brocker, err := initializeMesasgeQueue(ctx, logg, cfg)
	if err != nil {
		return err
	}

	// Initializing message queue connection.
	if err := brocker.Connect(ctx); err != nil {
		logg.Error(ctx, "connect message queue", slog.Any("err", err))
		return err
	}
	logg.Info(ctx, "message queue connection established")

	// Initializing the app.
	scheduler, err := initializeScheduler(ctx, logg, cfg, storage, brocker)
	if err != nil {
		return err
	}

	// Starting sending notifications.
	scheduler.StartProducer(ctx)
	logg.Info(ctx, "scheduler started successfully")

	// Starting periodic storage clean up.
	scheduler.StartCleanup(ctx)
	logg.Info(ctx, "scheduler started successfully")

	<-ctx.Done()
	scheduler.Wait(ctx)

	return nil
}

func loadConfig(ctx context.Context, logg *logger.Logger) (config.ServiceConfig, error) {
	loader := config.NewLoader(
		"scheduler",
		"Calendar scheduler service",
		"Scheduler service for managing notifications and periodic cleanup",
		defaultConfigFile,
		"CALENDAR",
	)
	cfg, err := loader.Load(&schedulerConfig.Config{}, printVersion, os.Stdout)
	if err != nil {
		if errors.Is(err, config.ErrShouldStop) {
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
	newLogg = newLogg.With(slog.String("service", "scheduler"))
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

func initializeScheduler(
	ctx context.Context,
	logg *logger.Logger,
	cfg config.ServiceConfig,
	storage storage.Storage,
	brocker *mq.RabbitMQ,
) (*schedulerPkg.Scheduler, error) {
	schCfg, err := cfg.GetSubConfig("app")
	if err != nil {
		logg.Error(ctx, "get scheduler app config", slog.Any("err", err))
		return nil, err
	}
	sch, err := schedulerPkg.NewScheduler(logg.With(slog.String("layer", "SCHEDULER")), storage, brocker, schCfg)
	if err != nil {
		logg.Error(ctx, "create scheduler", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "scheduler created successfully")
	return sch, nil
}

func initializeMesasgeQueue(
	ctx context.Context,
	logg *logger.Logger,
	cfg config.ServiceConfig,
) (*mq.RabbitMQ, error) {
	mqCfg, err := cfg.GetSubConfig("rmq")
	if err != nil {
		logg.Error(ctx, "get message queue config", slog.Any("err", err))
		return nil, err
	}
	brocker, err := mq.NewRabbitMQ(logg.With("layer", "RabbitMQ"), mqCfg, mq.ProducerOnly)
	if err != nil {
		logg.Error(ctx, "create message queue", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "message queue created successfully")
	return brocker, nil
}
