package task

import (
	"context"
)

type Task struct {
	id            string
	cfg           Config
	runF          func(context.Context) error
	shutDownF     func(context.Context) error
	state         *State
	notHandledErr chan<- Err
}

type Err struct {
	Err        error
	ID         string
	FallNumber int
}

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

func (t *Task) Run(ctx context.Context) error {
	for {
		if err := t.runF(ctx); err != nil {
			t.state.FallNumberInc()
			if t.cfg.FallNumberIsUnlimited() || t.state.GetFallNumber() <= t.cfg.GetFallNumber() {
				t.sendNotHandledErr(err)
				continue
			}
			return err
		}
		break
	}
	return nil
}
