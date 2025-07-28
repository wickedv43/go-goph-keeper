package kv

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
)

func setupTestKV(tb testing.TB) *KV {
	tb.Helper()

	tempDir := tb.TempDir()
	i := do.New()

	do.ProvideValue(i, &config.Config{KV: config.KV{DirPath: tempDir}})
	log, err := logger.NewLogger(i)
	require.NoError(tb, err)
	do.ProvideValue(i, log)

	kv, err := NewRoseDB(i)
	require.NoError(tb, err)
	tb.Cleanup(func() { kv.Shutdown() })

	return kv
}

func TestNewRoseDB(t *testing.T) {
	kv := setupTestKV(t)

	err := kv.db.Put([]byte("foo"), []byte("bar"))
	require.NoError(t, err)

	val, err := kv.db.Get([]byte("foo"))
	require.NoError(t, err)
	require.Equal(t, []byte("bar"), val)
}
