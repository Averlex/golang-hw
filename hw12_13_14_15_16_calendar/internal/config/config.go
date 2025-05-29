// Package config provides configuration structures for the service.
package config

import (
	"time"
)

// Config is a config for calendar service.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	Server  ServerConf  `mapstructure:"server"`
	App     AppConf     `mapstructure:"app"`
}

// LoggerConf is a config for logger.
type LoggerConf struct {
	Level        string `mapstructure:"level"`
	Format       string `mapstructure:"format"`
	TimeTemplate string `mapstructure:"time_template"`
}

// ServerConf is a config for the global server settings.
type ServerConf struct {
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// StorageConf is a config for storage containing storage type.
type StorageConf struct {
	Type    string        `mapstructure:"type"`
	Timeout time.Duration `mapstructure:"timeout"` // In seconds. 0 means timeout will be disabled.
	SQL     SQLConf       `mapstructure:"sql"`
	Memory  MemoryConf    `mapstructure:"memory"`
}

// SQLConf represents a database configuration used to build DSN string.
type SQLConf struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBname   string `mapstructure:"dbname"`
}

//nolint:revive,nolintlint
type MemoryConf struct {
	Size int `mapstructure:"size"`
}

// AppConf is a config for the global app settings, like environment (dev/prod) and log stream.
type AppConf struct {
	LogStream string `mapstructure:"log_stream"`
}

// ToMap converts StorageConf to a map[string]any.
//
// It returns a map containing all the fields of StorageConf as key-value pairs.
// The fields of SQLConf and MemoryConf are placed in sub-maps.
func (s *StorageConf) ToMap() map[string]any {
	return map[string]any{
		"type": s.Type,
		"sql": map[string]any{
			"host":     s.SQL.Host,
			"port":     s.SQL.Port,
			"user":     s.SQL.User,
			"password": s.SQL.Password,
			"dbname":   s.SQL.DBname,
			"timeout":  s.Timeout,
		},
		"memory": map[string]any{
			"size":    s.Memory.Size,
			"timeout": s.Timeout,
		},
	}
}
