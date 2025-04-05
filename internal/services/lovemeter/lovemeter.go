package lovemeter

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)
	Get(player string) int
}

type LoveMeterImpl struct {
	Values map[string]int
}

func NewLoveMeter() LoveMeter {
	return &LoveMeterImpl{Values: make(map[string]int)}
}

func (lm *LoveMeterImpl) Increase(player string, amount int) {
	lm.Values[player] += amount
	if lm.Values[player] > 100 {
		lm.Values[player] = 100
	}
}

func (lm *LoveMeterImpl) Decrease(player string, amount int) {
	lm.Values[player] -= amount
	if lm.Values[player] < 0 {
		lm.Values[player] = 0
	}
}

func (lm *LoveMeterImpl) Get(player string) int {
	return lm.Values[player]
}
