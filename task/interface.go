package task

import "context"

type Interface interface {
	Run(context.Context) error
	Shutdown(context.Context) error
	GetTaskConfig() Config
	GetID() string
}
