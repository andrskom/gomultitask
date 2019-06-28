package task

import (
	"context"
	"fmt"
	"time"
)

// Task is system wrapper for task
type Task struct {
	id            string
	cfg           Config
	runF          func(context.Context) error
	shutDownF     func(context.Context) error
	state         *State
	notHandledErr chan<- Err
}

// Err is internal task error
type Err struct {
	Err        error
	ID         string
	FallNumber int
}

// NewFromInterface build task wrapper from user's task
//  notHandledErr - is channel for send custom err while we can restart application
//  i - user's task
func NewFromInterface(notHandledErr chan<- Err, i Interface) *Task {
	return &Task{
		notHandledErr: notHandledErr,
		id:            i.GetID(),
		cfg:           i.GetTaskConfig(),
		runF:          i.Run,
		shutDownF:     i.Shutdown,
		state:         GetDefaultState(),
	}
}

func (t *Task) sendNotHandledErr(err error) {
	t.notHandledErr <- Err{
		ID:         t.id,
		FallNumber: t.state.GetFallNumber(),
		Err:        err,
	}
}

// Run task with restart while not reached fall limit or exit
func (t *Task) Run(ctx context.Context) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic in run: %+v", rec)
		}
	}()
	for {
		if err := t.runF(ctx); err != nil {
			t.state.FallNumberInc()
			if t.cfg.FallNumberIsUnlimited() || t.state.GetFallNumber() <= t.cfg.FallNumber {
				t.sendNotHandledErr(err)
				if t.cfg.HasRestartTimeout() {
					time.Sleep(t.cfg.RestartTimeout)
				}
				continue
			}
			t.state.SetFailed()
			return err
		}
		break
	}
	return nil
}

// Shutdown task, is failed, don't need to stop it
func (t *Task) Shutdown(ctx context.Context) error {
	if t.state.IsFailed() {
		return nil
	}
	return t.shutDownF(ctx)
}

// GetID return id of task
func (t *Task) GetID() string {
	return t.id
}
