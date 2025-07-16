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

	senderConfig "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/config/sender" //nolint:depguard
	senderPkg "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/sender"           //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/config"                          //nolint:depguard
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/logger"                          //nolint:depguard
	mq "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/rabbitmq"                     //nolint:depguard
)

const (
	exitCodeSuccess = 0
	exitCodeError   = 1
)

var defaultConfigFile = "../../configs/sender/config.toml"

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
	logg = logg.With(slog.String("service", "sender"))

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

	// Initializing signal handler.
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

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
	sender, err := initializeSender(ctx, logg, brocker)
	if err != nil {
		return err
	}

	err = sender.Start(ctx)
	if err != nil {
		return err
	}
	logg.Info(ctx, "sender started successfully")

	<-ctx.Done()
	sender.Wait(ctx)

	return nil
}

func loadConfig(ctx context.Context, logg *logger.Logger) (config.ServiceConfig, error) {
	loader := config.NewLoader(
		"sender",
		"Calendar sender service",
		"Sender service for managing notification sending",
		defaultConfigFile,
		"CALENDAR",
	)
	cfg, err := loader.Load(&senderConfig.Config{}, printVersion, os.Stdout)
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
	newLogg = newLogg.With(slog.String("service", "sender"))
	newLogg.Info(ctx, "logger created successfully")
	return newLogg, nil
}

func initializeSender(
	ctx context.Context,
	logg *logger.Logger,
	brocker *mq.RabbitMQ,
) (*senderPkg.Sender, error) {
	sch, err := senderPkg.NewSender(logg.With(slog.String("layer", "SENDER")), brocker)
	if err != nil {
		logg.Error(ctx, "create sender", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "sender created successfully")
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
	brocker, err := mq.NewRabbitMQ(logg.With("layer", "RabbitMQ"), mqCfg, mq.ConsumerOnly)
	if err != nil {
		logg.Error(ctx, "create message queue", slog.Any("err", err))
		return nil, err
	}
	logg.Info(ctx, "message queue created successfully")
	return brocker, nil
}
