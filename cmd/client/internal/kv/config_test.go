package kv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetAndGetConfig(t *testing.T) {
	kv := setupTestKV(t)

	cfg := Config{
		Current: "user1",
		Contexts: map[string]Context{
			"user1": {Token: "abc123", Key: "secret"},
		},
		ServerIP: "localhost",
	}

	require.NoError(t, kv.SetConfig(cfg))

	got, err := kv.GetConfig()
	require.NoError(t, err)
	require.Equal(t, cfg.Current, got.Current)
	require.Equal(t, cfg.Contexts["user1"], got.Contexts["user1"])
	require.Equal(t, cfg.ServerIP, got.ServerIP)

	cfg.Current = "user2"
	cfg.Contexts["user2"] = Context{Token: "def456", Key: "newsecret"}
	require.NoError(t, kv.SetConfig(cfg))

	got, err = kv.GetConfig()
	require.NoError(t, err)
	require.Equal(t, "user2", got.Current)
	require.Equal(t, "newsecret", got.Contexts["user2"].Key)
}

func TestSaveAndGetKey(t *testing.T) {
	kv := setupTestKV(t)

	require.NoError(t, kv.SaveContext("alice", ""))
	require.NoError(t, kv.SaveKey("alice", "top-secret"))

	key, err := kv.GetCurrentKey()
	require.NoError(t, err)
	require.Equal(t, "top-secret", key)
}

func TestSaveContextAndGetToken(t *testing.T) {
	kv := setupTestKV(t)

	// no context yet
	_, err := kv.GetCurrentToken()
	require.ErrorIs(t, err, ErrEmptyContext)

	// save token
	require.NoError(t, kv.SaveContext("bob", "token-xyz"))

	token, err := kv.GetCurrentToken()
	require.NoError(t, err)
	require.Equal(t, "token-xyz", token)

	// corrupted config: no context with current
	cfg, err := kv.GetConfig()
	require.NoError(t, err)
	cfg.Current = "ghost"
	require.NoError(t, kv.SetConfig(cfg))

	_, err = kv.GetCurrentToken()
	require.ErrorIs(t, err, ErrEmptyContext)
}

func TestUseContext(t *testing.T) {
	kv := setupTestKV(t)

	// unknown context
	err := kv.UseContext("ghost")
	require.ErrorIs(t, err, ErrContextNotFound)

	// valid context
	require.NoError(t, kv.SaveContext("userA", "tokenA"))
	require.NoError(t, kv.SaveKey("userA", "keyA"))

	require.NoError(t, kv.UseContext("userA"))

	key, err := kv.GetCurrentKey()
	require.NoError(t, err)
	require.Equal(t, "keyA", key)

	token, err := kv.GetCurrentToken()
	require.NoError(t, err)
	require.Equal(t, "tokenA", token)
}
