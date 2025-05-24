package main

import "time"

// Config is a config for calendar service.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	Server  ServerConf  `mapstructure:"server"`
	App     AppConf     `mapstructure:"app"`
}

// LoggerConf is a config for logger.
type LoggerConf struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// ServerConf is a config for the global server settings.
type ServerConf struct {
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// StorageConf is a config for storage containing storage type.
type StorageConf struct {
	Type string `mapstructure:"type"`
}

// AppConf is a config for the global app settings, like environment (dev/prod) and log stream.
type AppConf struct {
	Env       string `mapstructure:"env"`
	LogStream string `mapstructure:"log_stream"`
}
