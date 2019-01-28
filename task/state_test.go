package task

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultState(t *testing.T) {
	r := require.New(t)

	state := GetDefaultState()
	r.NotNil(state)
	r.Equal(false, state.IsFailed())
	r.Equal(0, state.GetFallNumber())
}

func TestState_FallNumberInc(t *testing.T) {
	r := require.New(t)

	state := GetDefaultState()
	eVal := 10
	for i := 0; i < eVal; i++ {
		state.FallNumberInc()
	}
	r.Equal(eVal, state.GetFallNumber())
}

func TestState_IsFailed(t *testing.T) {
	r := require.New(t)

	state := GetDefaultState()
	r.False(state.IsFailed())
	state.SetFailed()
	r.True(state.IsFailed())
}
