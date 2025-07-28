package kv

import (
	"encoding/json"

	"github.com/pkg/errors"
)

var (
	ErrEmptyKey        = errors.New("empty key")
	ErrEmptyContext    = errors.New("empty context")
	ErrContextNotFound = errors.New("context not found")
)

const nsConfig = "config:"

type Context struct {
	Token string `json:"token"`
	Key   string `json:"key"`
}

type Config struct {
	Current  string             `json:"current"`
	Contexts map[string]Context `json:"contexts"`
	ServerIP string             `json:"server_ip"`
}

func (s *KV) SetConfig(cfg Config) error {
	keyByte := []byte(nsConfig)
	valByte, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "marshal account")
	}

	err = s.db.Put(keyByte, valByte)
	if err != nil {
		return errors.Wrap(err, "put kv")
	}

	return nil
}

func (s *KV) GetConfig() (Config, error) {
	keyByte := []byte(nsConfig)
	val, err := s.db.Get(keyByte)
	if err != nil {
		return Config{}, errors.Wrap(err, "put kv")
	}
	var c Config
	err = json.Unmarshal(val, &c)
	if err != nil {
		return Config{}, errors.Wrap(err, "json unmarshal failed")
	}

	return c, nil
}

func (s *KV) SaveContext(login, token string) error {
	cfg, err := s.GetConfig()
	if err != nil {
		var c Config
		c.Contexts = make(map[string]Context)
		cfg = c
	}
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}
	key, err := s.GetCurrentKey()
	if err != nil {
		key = ""
	}

	cfg.Contexts[login] = Context{Token: token, Key: key}
	cfg.Current = login
	return s.SetConfig(cfg)
}

func (s *KV) SaveKey(login, key string) error {
	cfg, err := s.GetConfig()
	if err != nil {
		var c Config
		c.Contexts = make(map[string]Context)
		cfg = c
	}
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}
	token, err := s.GetCurrentToken()
	if err != nil {
		token = ""
	}

	cfg.Contexts[login] = Context{Token: token, Key: key}
	cfg.Current = login
	return s.SetConfig(cfg)
}

func (s *KV) UseContext(name string) error {
	cfg, _ := s.GetConfig()
	if _, ok := cfg.Contexts[name]; !ok {
		return ErrContextNotFound
	}
	cfg.Current = name
	return s.SetConfig(cfg)
}

func (s *KV) GetCurrentToken() (string, error) {
	cfg, _ := s.GetConfig()
	if ctx, ok := cfg.Contexts[cfg.Current]; ok {
		return ctx.Token, nil
	}
	return "", ErrEmptyContext
}

func (s *KV) GetCurrentKey() (string, error) {
	cfg, _ := s.GetConfig()
	if ctx, ok := cfg.Contexts[cfg.Current]; ok {
		if ctx.Key == "" {
			return "", ErrEmptyKey
		}

		return ctx.Key, nil
	}
	return "", ErrContextNotFound
}
