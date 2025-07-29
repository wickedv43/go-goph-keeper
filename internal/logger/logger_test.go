package logger

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	i := do.New()

	l, err := NewLogger(i)
	require.NoError(t, err)
	require.NotNil(t, l)

	// Дополнительно проверим, что можно что-то логировать
	require.NotPanics(t, func() {
		l.Infof("test log message")
	})
}
