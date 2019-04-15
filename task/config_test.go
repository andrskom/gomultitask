package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfig(t *testing.T) {
	r := require.New(t)

	cfg := GetDefaultConfig()
	r.Equal(0, cfg.FallNumber)
	r.Equal(time.Duration(0), cfg.RestartTimeout)
}

func TestConfig_FallNumberIsUnlimited(t *testing.T) {
	r := require.New(t)

	cfg := GetDefaultConfig()
	cfg.FallNumber = -1
	r.True(cfg.FallNumberIsUnlimited())
	cfg.FallNumber = 0
	r.False(cfg.FallNumberIsUnlimited())
}

func TestConfig_HasRestartTimeout(t *testing.T) {
	r := require.New(t)

	cfg := GetDefaultConfig()
	r.False(cfg.HasRestartTimeout())
	cfg.RestartTimeout = 1
	r.True(cfg.HasRestartTimeout())
}
