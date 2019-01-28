package gomultitask

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrskom/gomultitask/task"
)

type Logger interface {
	Infof(string, ...interface{})
	Errorf(string, ...interface{})
}

type Operator struct {
	log             Logger
	tasks           []*task.Task
	notHandledErr   chan task.Err
	shutdownSignals []os.Signal
}

func NewOperator(list ...task.Interface) *Operator {
	taskList := make([]*task.Task, 0)
	for _, t := range list {
		taskList = append(taskList, task.NewFromInterface(t))
	}
	return &Operator{
		tasks:           taskList,
		notHandledErr:   make(chan task.Err, 5),
		shutdownSignals: []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT},
	}
}

func (o *Operator) WithLogger(log Logger) *Operator {
	o.log = log
	return o
}

func (o *Operator) Run(ctx context.Context) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, o.shutdownSignals...)

	internalCtx, cancelF := context.WithCancel(ctx)
	// init background notHandledErr logger
	go o.logNotHandledErr(internalCtx)

	//run all tasks
	for _, t := range o.tasks {
		go func() {
			if err := t.Run(ctx); err != nil {
				// TODO
			}
		}()
	}

	sig := <-sigCh

	return nil
}

func (o *Operator) logNotHandledErr(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case err := <-o.notHandledErr:
		if o.log != nil {
			o.log.Errorf("ID: %s, FallNumber: %d, Err: %s", err.ID, err.FallNumber, err.Err.Error())
		}
	}
}
