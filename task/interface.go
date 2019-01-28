package task

import "context"

//Interface of run in system task
type Interface interface {
	//run task function
	Run(context.Context) error
	//shutdown task function
	Shutdown(context.Context) error
	//return config of task
	GetTaskConfig() Config
	//return id of task, it will be use for beauty logs
	GetID() string
}
