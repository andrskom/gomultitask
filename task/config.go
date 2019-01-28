package task

type Config struct {
	// <  0 - unlimited(always restart)
	// == 0 - stop all after first err
	// >  0 - stop after fail number > FallNumber
	// default 0
	FallNumber *int
}

func (tc *Config) GetFallNumber() int {
	if tc.FallNumber == nil {
		return 0
	}
	return *tc.FallNumber
}

func (tc *Config) FallNumberIsUnlimited() bool {
	if tc.FallNumber == nil {
		return false
	}
	return *tc.FallNumber < 0
}

