package config

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`

	Envinronment string `mapstructure:"envinronment"`
}

type Server struct {
	Port string `mapstructure:"port"`
}

type Database struct {
	DSN string `mapstructure:"dsn"`
}

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

	return &cfg, nil
}
