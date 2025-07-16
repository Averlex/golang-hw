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
	HTTP    HTTPConf    `mapstructure:"http"`
	GRPC    GRPCConf    `mapstructure:"grpc"`
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

// MemoryConf is a config for memory storage.
type MemoryConf struct {
	Size int `mapstructure:"size"`
}

// AppConf is a config for the global app settings, like retry timeout and number of retries.
type AppConf struct {
	RetryTimeout time.Duration `mapstructure:"retry_timeout"`
	Retries      int           `mapstructure:"retries"`
}

// HTTPConf is a config for http server.
type HTTPConf struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// GRPCConf is a config for gRPC server.
type GRPCConf struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}
