package task

import (
	"time"
)

//Config for task
type Config struct {
	// <  0 - unlimited(always restart)
	// == 0 - stop all after first err
	// >  0 - stop after fail number > FallNumber
	FallNumber int
	// <= 0 - use like no timeout
	RestartTimeout time.Duration
}

//GetDefaultConfig return default set config
func GetDefaultConfig() Config {
	return Config{
		FallNumber:     0,
		RestartTimeout: 0,
	}
}

//FallNumberIsUnlimited return true for not limited fall for task
func (tc *Config) FallNumberIsUnlimited() bool {
	return tc.FallNumber < 0
}

//HasRestartTimeout return info about timeout
func (tc *Config) HasRestartTimeout() bool {
	return tc.RestartTimeout > 0
}