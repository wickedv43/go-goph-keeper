package kv

type Storage interface {
	SetConfig(cfg Config) error
	GetConfig() (Config, error)

	SaveContext(login, token string) error
	SaveKey(login, key string) error
	UseContext(name string) error

	GetCurrentToken() (string, error)
	GetCurrentKey() (string, error)
}
