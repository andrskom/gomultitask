package task

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	r := require.New(t)

	cfg := GetDefaultConfig()
	r.Equal(0, cfg.FallNumber)
}

func TestConfig_FallNumberIsUnlimited(t *testing.T) {
	r := require.New(t)

	cfg := GetDefaultConfig()
	cfg.FallNumber = -1
	r.True(cfg.FallNumberIsUnlimited())
}
