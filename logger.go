package gomultitask

type Logger interface {
	Infof(string, ...interface{})
	Errorf(string, ...interface{})
}

func (o *Operator) logInfof(msg string, args ...interface{}) {
	if o.log != nil {
		o.log.Infof(msg, args...)
	}
}

func (o *Operator) logErrorf(msg string, args ...interface{}) {
	if o.log != nil {
		o.log.Errorf(msg, args...)
	}
}
