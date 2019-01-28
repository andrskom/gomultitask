package task

//Config for task
type Config struct {
	// <  0 - unlimited(always restart)
	// == 0 - stop all after first err
	// >  0 - stop after fail number > FallNumber
	FallNumber int
}

//GetDefaultConfig return default set config
func GetDefaultConfig() Config {
	return Config{
		FallNumber: 0,
	}
}

//FallNumberIsUnlimited return true for not limited fall for task
func (tc *Config) FallNumberIsUnlimited() bool {
	return tc.FallNumber < 0
}
