package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGophKeeper_LoginCMD(t *testing.T) {
	g := setupTestClient(t)

	err := g.LoginCMD().RunE(g.rootCmd, nil)
	require.Error(t, err)
}

func TestGophKeeper_RegisterCMD(t *testing.T) {
	g := setupTestClient(t)

	err := g.RegisterCMD().RunE(g.rootCmd, nil)
	require.Error(t, err)
}

func TestGophKeeper_NewVaultCMD(t *testing.T) {
	g := setupTestClient(t)

	err := g.NewVaultCMD().RunE(g.rootCmd, nil)
	require.Error(t, err)
}

func TestGophKeeper_VaultListCMD(t *testing.T) {
	g := setupTestClient(t)

	err := g.VaultListCMD().RunE(g.rootCmd, nil)
	require.Error(t, err)
}
