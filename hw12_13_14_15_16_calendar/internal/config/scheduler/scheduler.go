// Package scheduler provides configuration structures for the service.
package scheduler

import (
	"time"
)

// Config is a config for calendar service.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	RMQ     RMQConf     `mapstructure:"rmq"`
	App     AppConf     `mapstructure:"app"`
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

// RMQConf is a config for Rabbit MQ client.
type RMQConf struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	Timeout      time.Duration `mapstructure:"timeout"`
	RetryTimeout time.Duration `mapstructure:"retry_timeout"`
	Retries      int           `mapstructure:"retries"`
	Topic        string        `mapstructure:"topic"`
	Durable      bool          `mapstructure:"durable"`
	ContentType  string        `mapstructure:"content_type"`
	RoutingKey   string        `mapstructure:"routing_key"`
}

// AppConf is a config for the global app settings, like retry timeout and number of retries,
// as well as queue intervals.
type AppConf struct {
	RetryTimeout    time.Duration `mapstructure:"retry_timeout"`
	Retries         int           `mapstructure:"retries"`
	QueueInterval   time.Duration `mapstructure:"queue_interval"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}
