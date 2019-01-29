package gomultitask

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/andrskom/gomultitask/task"
	"github.com/stretchr/testify/require"
)

type TestingTask struct {
	id              string
	cfg             task.Config
	finishTaskCh    chan error
	panicTask       chan string
	shutdownTimeout time.Duration
	shutdownErr     error
}

func NewTestingTask(id string, cfg task.Config, shutdownTimeout time.Duration) *TestingTask {
	return &TestingTask{
		id:              id,
		cfg:             cfg,
		shutdownTimeout: shutdownTimeout,
		finishTaskCh:    make(chan error),
		panicTask:       make(chan string),
	}
}

func (t *TestingTask) Run(context.Context) error {
	for {
		select {
		case err := <-t.finishTaskCh:
			return err
		case msg := <-t.panicTask:
			panic(msg)
		}
	}
}

func (t *TestingTask) Shutdown(context.Context) error {
	time.Sleep(t.shutdownTimeout)
	return t.shutdownErr
}

func (t *TestingTask) GetTaskConfig() task.Config {
	return t.cfg
}

func (t *TestingTask) GetID() string {
	return t.id
}

type TestingLogger struct {
	errChan  chan string
	infoChan chan string
}

func NewTestingLogger() *TestingLogger {
	return &TestingLogger{
		errChan:  make(chan string, 10),
		infoChan: make(chan string, 10),
	}
}

func (l *TestingLogger) Infof(msg string, args ...interface{}) {
	l.infoChan <- fmt.Sprintf(msg, args...)
}

func (l *TestingLogger) Errorf(msg string, args ...interface{}) {
	l.errChan <- fmt.Sprintf(msg, args...)
}

func getPreparedTask(t *testing.T, num int) []*TestingTask {
	require.True(t, num > 0)
	res := make([]*TestingTask, 0)
	for i := 0; i < num; i++ {
		res = append(res, NewTestingTask("testingTask"+strconv.Itoa(i), task.Config{FallNumber: i}, 0))
	}
	return res
}

func TestBaseConfiguration(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 2)
	op := NewOperator(tasks[0], tasks[1])
	r.Equal(defaultShutdownDeadline, op.shutdownDeadline)
	r.Nil(op.log)
	r.Equal([]os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT}, op.shutdownSignals)
	r.Len(op.tasks, 2)

	eLogger := NewTestingLogger()
	op.WithLogger(eLogger)
	r.NotNil(op.log)
	r.Equal(eLogger, op.log)

	eDeadline := 5 * time.Second
	op.WithShutdownDeadline(eDeadline)
	r.Equal(eDeadline, op.shutdownDeadline)

	eSignals := []os.Signal{syscall.SIGUSR1}
	op.WithShutdownSignals(eSignals)
	r.Equal(eSignals, op.shutdownSignals)
}

func TestErrLogging(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	tLogger := NewTestingLogger()
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4]).WithLogger(tLogger)
	tCh := make(chan error)
	go func() {
		tCh <- op.Run(context.Background())
	}()

	errMsgList := make([]string, 0)
	for i := 1; i <= 3; i++ {
		errMsgList = append(
			errMsgList,
			fmt.Sprintf("ID: testingTask%d, FallNumber: 1, Err: expected err %d", i, i),
		)
		tasks[i].finishTaskCh <- fmt.Errorf("expected err %d", i)
	}
	for i := 1; i <= 3; i++ {
		select {
		case msg := <-tLogger.errChan:
			r.Contains(errMsgList, msg)
		case <-time.After(time.Second):
			r.Fail("Not get errs in time")
		}
	}

	go func() {
		op.sigCh <- syscall.SIGTERM
	}()
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(time.Second):
		r.Fail("Not shutdowned in expected time")
	}
}

func TestCorrectShutdown(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4])
	tCh := make(chan error)
	go func() {
		tCh <- op.Run(context.Background())
	}()
	go func() {
		op.sigCh <- syscall.SIGTERM
	}()
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(1 * time.Second):
		r.Fail("Not shutdowned in expected time")
	}
}

func TestCorrectShutdownAfterSomeFail(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4])
	tCh := make(chan error)
	go func() {
		tCh <- op.Run(context.Background())
	}()
	tasks[2].finishTaskCh <- errors.New("expected error")
	go func() {
		op.sigCh <- syscall.SIGTERM
	}()
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(1 * time.Second):
		r.Fail("Not shutdowned in expected time")
	}
}

func TestCorrectShutdownByGroupStop(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4])
	tCh := make(chan error)
	go func() {
		tCh <- op.Run(context.Background())
	}()
	tasks[0].finishTaskCh <- errors.New("expected error")
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(1 * time.Second):
		r.Fail("Not shutdowned in expected time")
	}
}

func TestCorrectShutdownByPanic(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4])
	tCh := make(chan error)
	go func() {
		tCh <- op.Run(context.Background())
	}()
	tasks[0].panicTask <- "expected panic"
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(1 * time.Second):
		r.Fail("Not shutdowned in expected time")
	}
}

func TestShutdownErr(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	eErr := errors.New("expected error")
	tasks[3].shutdownErr = eErr
	tLogger := NewTestingLogger()
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4]).WithLogger(tLogger)
	tCh := make(chan error)

	go func() {
		tCh <- op.Run(context.Background())
	}()
	go func() {
		op.sigCh <- syscall.SIGTERM
	}()
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(1 * time.Second):
		r.Fail("Not shutdowned in expected time")
	}
	foundShutdownErr := false
TestWaiter:
	for {
		select {
		case msg := <-tLogger.errChan:
			if msg == "Have got 1 errors while shutdown tasks" {
				foundShutdownErr = true
				break TestWaiter
			}
		case <-time.After(time.Second):
			r.Fail("have not err int time")
		}
	}
	r.True(foundShutdownErr)
}

func TestShutdownDeadline(t *testing.T) {
	r := require.New(t)

	tasks := getPreparedTask(t, 5)
	tasks[3].shutdownTimeout = 5 * time.Second
	tLogger := NewTestingLogger()
	op := NewOperator(tasks[0], tasks[1], tasks[2], tasks[3], tasks[4]).
		WithLogger(tLogger).
		WithShutdownDeadline(time.Second)
	tCh := make(chan error)

	go func() {
		tCh <- op.Run(context.Background())
	}()
	go func() {
		op.sigCh <- syscall.SIGTERM
	}()
	select {
	case err := <-tCh:
		r.NoError(err)
	case <-time.After(2 * time.Second):
		r.Fail("Not shutdowned in expected time")
	}
	foundShutdownErr := false
TestWaiter:
	for {
		select {
		case msg := <-tLogger.errChan:
			if msg == "Deadline for graceful shutdown is reached" {
				foundShutdownErr = true
				break TestWaiter
			}
		case <-time.After(time.Second):
			r.Fail("have not err int time")
		}
	}
	r.True(foundShutdownErr)
}
