package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShellCMD(t *testing.T) {
	g := setupTestClient(t)

	cmd := g.ShellCMD()

	require.Equal(t, "shell", cmd.Use)

	err := cmd.RunE(cmd, nil)
	require.NoError(t, err)
}

func TestGophKeeper_shellLoop(t *testing.T) {
	g := setupTestClient(t)

	err := g.shellLoop()
	require.NoError(t, err)

}
