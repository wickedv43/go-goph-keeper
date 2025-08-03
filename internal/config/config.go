// Package config provides application configuration loading and parsing.
package config

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// Config holds the full application configuration loaded from file.
type Config struct {
	Server       Server   `mapstructure:"server"`
	Database     Database `mapstructure:"database"`
	KV           KV       `mapstructure:"databaseKV"`
	Master       string
	Envinronment string `mapstructure:"envinronment"`
}

// Server contains server-related configuration such as port.
type Server struct {
	Port string `mapstructure:"port"`
}

// Database contains database connection settings.
type Database struct {
	DSN string `mapstructure:"dsn"`
}

// KV contains configuration for key-value storage.
type KV struct {
	DirPath string `mapstructure:"dirPath"`
}

// NewConfig loads configuration from a file using viper and sets defaults where needed.
func NewConfig(i do.Injector) (*Config, error) {
	configPath := do.MustInvokeNamed[string](i, "config.path")
	dir, file := filepath.Split(configPath)
	nameParts := strings.Split(file, ".")
	if len(nameParts) != 3 {
		return nil, errors.Errorf("invalid config file name: %s", file)
	}

	name, ext, enviroment := strings.Join(nameParts[:2], "."), nameParts[2], nameParts[1]

	viper.SetConfigName(name)
	viper.SetConfigType(ext)
	viper.AddConfigPath(dir)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "read config")
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode into struct")
	}

	if cfg.Envinronment == "" {
		cfg.Envinronment = enviroment
	}

	cfg.Master = viper.GetString("master")
	return &cfg, nil
}
