package task

type State struct {
	fallNumber int
}

func GetDefaultState() *State {
	return &State{}
}

func (s *State) FallNumberInc() {
	s.fallNumber++
}

func (s *State) GetFallNumber() int {
	return s.fallNumber
}
