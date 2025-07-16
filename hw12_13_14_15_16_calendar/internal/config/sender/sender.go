// Package sender provides configuration structures for the service.
package sender

import (
	"time"
)

// Config is a config for calendar service.
type Config struct {
	Logger LoggerConf `mapstructure:"logger"`
	RMQ    RMQConf    `mapstructure:"rmq"`
}

// LoggerConf is a config for logger.
type LoggerConf struct {
	Level        string `mapstructure:"level"`
	Format       string `mapstructure:"format"`
	TimeTemplate string `mapstructure:"time_template"`
	LogStream    string `mapstructure:"log_stream"`
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
	ResubTimeout time.Duration `mapstructure:"resub_timeout"`
	AutoAck      bool          `mapstructure:"auto_ack"`
	Requeue      bool          `mapstructure:"requeue"`
}
