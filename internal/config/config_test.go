package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Run("success: valid config file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.dev.yaml")

		err := os.WriteFile(path, []byte(`
server:
  port: "8080"
database:
  dsn: "postgres://localhost/db"
databaseKV:
  dirPath: "/tmp/kv"
master: "admin"
`), 0644)
		require.NoError(t, err)

		i := do.New()
		do.ProvideNamedValue[string](i, "config.path", path)

		cfg, err := NewConfig(i)
		require.NoError(t, err)
		require.Equal(t, "8080", cfg.Server.Port)
		require.Equal(t, "postgres://localhost/db", cfg.Database.DSN)
		require.Equal(t, "/tmp/kv", cfg.KV.DirPath)
		require.Equal(t, "admin", cfg.Master)
		require.Equal(t, "dev", cfg.Envinronment)
	})

	t.Run("error: bad config file name (no 3 parts)", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")

		_ = os.WriteFile(path, []byte(``), 0644)

		i := do.New()
		do.ProvideNamedValue[string](i, "config.path", path)

		_, err := NewConfig(i)
		require.ErrorContains(t, err, "invalid config file name")
	})

	t.Run("error: missing config file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.dev.yaml") // не создаём файл

		i := do.New()
		do.ProvideNamedValue[string](i, "config.path", path)

		_, err := NewConfig(i)
		require.ErrorContains(t, err, "read config")
	})

	t.Run("error: malformed YAML structure", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.dev.yaml")

		// Здесь мы делаем валидный YAML, но с неправильной структурой
		err := os.WriteFile(path, []byte(`
server: "not a map"
`), 0644)
		require.NoError(t, err)

		i := do.New()
		do.ProvideNamedValue[string](i, "config.path", path)

		_, err = NewConfig(i)
		require.ErrorContains(t, err, "unable to decode")
	})
}
