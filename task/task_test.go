package task

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TaskMock struct {
	mock.Mock
	panicOnRun bool
}

func (m *TaskMock) Run(ctx context.Context) error {
	args := m.Called(ctx)
	if m.panicOnRun {
		panic("expected panic")
	}
	return args.Error(0)
}

func (m *TaskMock) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *TaskMock) GetTaskConfig() Config {
	args := m.Called()
	return args.Get(0).(Config)
}

func (m *TaskMock) GetID() string {
	args := m.Called()
	return args.String(0)
}

func TestNewFromInterface(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err)
	m := &TaskMock{}
	expectedID := "expectedID"
	m.On("GetID").Return(expectedID)
	expectedConfig := GetDefaultConfig()
	m.On("GetTaskConfig").Return(expectedConfig)
	task := NewFromInterface(ch, m)
	r.Equal(expectedID, task.GetID())
	r.Equal(expectedConfig, task.cfg)
}

func TestTask_Run_Finished(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{}
	m.On("GetID").Return("expectedID")
	m.On("GetTaskConfig").Return(GetDefaultConfig())
	m.On("Run", mock.Anything).Return(nil)
	task := NewFromInterface(ch, m)
	err := task.Run(context.Background())
	r.NoError(err)
	r.Len(ch, 0)
}

func TestTask_Run_FinishedErr(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{}
	m.On("GetID").Return("expectedID")
	m.On("GetTaskConfig").Return(GetDefaultConfig())
	eErr := errors.New("expected error")
	m.On("Run", mock.Anything).Return(eErr)
	task := NewFromInterface(ch, m)
	err := task.Run(context.Background())
	r.Error(err)
	r.Len(ch, 0)
	r.Equal(eErr, err)
}

func TestTask_Run_FinishedNotReachedErrLimit(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{}
	m.On("GetID").Return("expectedID")
	cfg := GetDefaultConfig()
	cfg.FallNumber = 3
	m.On("GetTaskConfig").Return(cfg)
	eErr := Err{
		ID:  "expectedID",
		Err: errors.New("expected err"),
	}
	m.On("Run", mock.Anything).Return(eErr.Err).Times(cfg.FallNumber)
	m.On("Run", mock.Anything).Return(nil)
	task := NewFromInterface(ch, m)
	err := task.Run(context.Background())
	r.NoError(err)
	r.Len(ch, cfg.FallNumber)
	for i := 0; i < cfg.FallNumber; i++ {
		err := <-ch
		eErr.FallNumber++
		r.Equal(eErr, err)
	}
}

func TestTask_Run_FinishedReachedErrLimit(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{}
	m.On("GetID").Return("expectedID")
	cfg := GetDefaultConfig()
	cfg.FallNumber = 3
	m.On("GetTaskConfig").Return(cfg)
	eErr := Err{
		ID:  "expectedID",
		Err: errors.New("expected err"),
	}
	m.On("Run", mock.Anything).Return(eErr.Err).Times(cfg.FallNumber + 1)
	task := NewFromInterface(ch, m)
	err := task.Run(context.Background())
	r.Error(err)
	r.Equal(eErr.Err, err)
	r.Len(ch, cfg.FallNumber)
	for i := 0; i < cfg.FallNumber; i++ {
		err := <-ch
		eErr.FallNumber++
		r.Equal(eErr, err)
	}
}

func TestTask_Run_Panic(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{panicOnRun: true}
	m.On("GetID").Return("expectedID")
	cfg := GetDefaultConfig()
	cfg.FallNumber = 3
	m.On("GetTaskConfig").Return(cfg)
	m.On("Run", mock.Anything).Return(nil)
	task := NewFromInterface(ch, m)
	err := task.Run(context.Background())
	r.Error(err)
	r.Contains(err.Error(), "panic")
}

func TestTask_Shutdown(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{panicOnRun: true}
	m.On("GetID").Return("expectedID")
	cfg := GetDefaultConfig()
	cfg.FallNumber = 3
	m.On("GetTaskConfig").Return(cfg)
	m.On("Shutdown", mock.Anything).Return(nil)
	task := NewFromInterface(ch, m)
	err := task.Shutdown(context.Background())
	r.NoError(err)
}

func TestTask_Shutdown_Err(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{panicOnRun: true}
	m.On("GetID").Return("expectedID")
	cfg := GetDefaultConfig()
	cfg.FallNumber = 3
	m.On("GetTaskConfig").Return(cfg)
	eErr := errors.New("expected error")
	m.On("Shutdown", mock.Anything).Return(eErr)
	task := NewFromInterface(ch, m)
	err := task.Shutdown(context.Background())
	r.Error(err)
	r.Equal(eErr, err)
}

func TestTask_ShutdownFailed(t *testing.T) {
	r := require.New(t)

	ch := make(chan Err, 10)
	m := &TaskMock{panicOnRun: true}
	m.On("GetID").Return("expectedID")
	cfg := GetDefaultConfig()
	cfg.FallNumber = 3
	m.On("GetTaskConfig").Return(cfg)
	eErr := errors.New("expected error")
	m.On("Shutdown", mock.Anything).Return(eErr)
	task := NewFromInterface(ch, m)
	task.state.failed = true
	err := task.Shutdown(context.Background())
	r.NoError(err)
}
