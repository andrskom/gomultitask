package gomultitask

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/andrskom/gomultitask/task"
)

const defaultShutdownDeadline = 30 * time.Second

//Operator is main struct for managing tasks
type Operator struct {
	log              Logger
	tasks            []*task.Task
	notHandledErr    chan task.Err
	sigCh            chan os.Signal
	errCh            chan error
	quitCh           chan struct{}
	shutdownSignals  []os.Signal
	shutdownDeadline time.Duration
}

//NewOperator init default operator for tasks
func NewOperator(list ...task.Interface) *Operator {
	taskList := make([]*task.Task, 0)
	notHandledErr := make(chan task.Err, 5)
	for _, t := range list {
		taskList = append(taskList, task.NewFromInterface(notHandledErr, t))
	}

	return &Operator{
		tasks:            taskList,
		notHandledErr:    notHandledErr,
		sigCh:            make(chan os.Signal, 1),
		errCh:            make(chan error),
		quitCh:           make(chan struct{}),
		shutdownSignals:  []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT},
		shutdownDeadline: defaultShutdownDeadline,
	}
}

//WithLogger add logger
func (o *Operator) WithLogger(log Logger) *Operator {
	o.log = log
	return o
}

//WithShutdownDeadline add shutdown deadline, default is 30s
func (o *Operator) WithShutdownDeadline(duration time.Duration) *Operator {
	o.shutdownDeadline = duration
	return o
}

//WithShutdownSignals set custom shutdown signals
func (o *Operator) WithShutdownSignals(signals []os.Signal) *Operator {
	o.shutdownSignals = signals
	return o
}

//Run tasks and wait while stop
func (o *Operator) Run(ctx context.Context) error {
	//signals catcher
	signal.Notify(o.sigCh, o.shutdownSignals...)

	// internal context for supply routines
	internalCtx, cancelF := context.WithCancel(ctx)
	defer cancelF()
	// init background notHandledErr logger
	go o.logNotHandledErr(internalCtx)

	//run all tasks
	for _, t := range o.tasks {
		go func(t *task.Task) {
			if err := t.Run(ctx); err != nil {
				o.errCh <- err
			}
		}(t)
	}

	// wait signal or error group
	go o.waitEnd(context.Background())

	// wait end of graceful shutdown
	<-o.quitCh

	return nil
}

func (o *Operator) waitEnd(ctx context.Context) {
	select {
	case sig := <-o.sigCh:
		o.logInfof("Signal caught: %s", sig.String())
		o.shutdown(ctx)
	case err := <-o.errCh:
		o.logErrorf("Error in group caught: %s", err.Error())
		o.shutdown(ctx)
	}
}

func (o *Operator) shutdown(ctx context.Context) {
	var wg sync.WaitGroup
	var shutdownErrCount int64
	for _, t := range o.tasks {
		wg.Add(1)
		go func(t *task.Task) {
			defer wg.Done()
			if err := t.Shutdown(ctx); err != nil {
				atomic.AddInt64(&shutdownErrCount, 1)
				o.logErrorf("Shutdown task ID %s, err %s", t.GetID(), err.Error())
			}
		}(t)
	}
	shutdownFinishedCH := make(chan struct{})
	go func() {
		wg.Wait()
		shutdownFinishedCH <- struct{}{}
	}()
	select {
	case <-shutdownFinishedCH:
		if shutdownErrCount > 0 {
			o.logErrorf("Have got %d errors while shutdown tasks", shutdownErrCount)
			break
		}
		o.logInfof("All graceful shutdowned")
	case <-time.After(o.shutdownDeadline):
		o.logErrorf("Deadline for graceful shutdown is reached")
	}
	o.quitCh <- struct{}{}
}

func (o *Operator) logNotHandledErr(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-o.notHandledErr:
			o.logErrorf("ID: %s, FallNumber: %d, Err: %s", err.ID, err.FallNumber, err.Err.Error())
		}
	}
}
