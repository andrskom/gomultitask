package task

// State of task
type State struct {
	fallNumber int
	failed     bool
}

// GetDefaultState build default state
func GetDefaultState() *State {
	return &State{
		failed:     false,
		fallNumber: 0,
	}
}

// FallNumberInc add one fall to state
func (s *State) FallNumberInc() {
	s.fallNumber++
}

// GetFallNumber return falls number
func (s *State) GetFallNumber() int {
	return s.fallNumber
}

// SetFailed register that task was failed and don't need shutdown it
func (s *State) SetFailed() {
	s.failed = true
}

// IsFailed return failed status
func (s *State) IsFailed() bool {
	return s.failed
}
