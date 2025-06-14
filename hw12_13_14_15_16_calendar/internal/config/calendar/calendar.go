// Package calendar provides configuration structures for the service.
package calendar

import (
	"time"
)

// Config is a config for calendar service.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	App     AppConf     `mapstructure:"app"`
	Server  ServerConf  `mapstructure:"server"`
}

// LoggerConf is a config for logger.
type LoggerConf struct {
	Level        string `mapstructure:"level"`
	Format       string `mapstructure:"format"`
	TimeTemplate string `mapstructure:"time_template"`
	LogStream    string `mapstructure:"log_stream"`
}

// StorageConf is a config for storage containing storage type.
type StorageConf struct {
	Type   string     `mapstructure:"type"`
	SQL    SQLConf    `mapstructure:"sql"`
	Memory MemoryConf `mapstructure:"memory"`
}

// SQLConf represents a database configuration used to build DSN string.
type SQLConf struct {
	Host     string        `mapstructure:"host"`
	Port     string        `mapstructure:"port"`
	User     string        `mapstructure:"user"`
	Password string        `mapstructure:"password"`
	DBname   string        `mapstructure:"dbname"`
	Timeout  time.Duration `mapstructure:"timeout"` // 0 means timeout will be disabled.
	Driver   string        `mapstructure:"driver"`
}

//nolint:revive,nolintlint
type MemoryConf struct {
	Size int `mapstructure:"size"`
}

// AppConf is a config for the global app settings, like environment (dev/prod) and log stream.
type AppConf struct {
	RetryTimeout time.Duration `mapstructure:"retry_timeout"`
	Retries      int           `mapstructure:"retries"`
}

// ServerConf is a config for the global server settings.
type ServerConf struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}
