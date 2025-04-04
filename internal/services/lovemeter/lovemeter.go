package lovemeter

type LoveMeter interface {
	Increase(amount int)
	Decrease(amount int)
	Get() int
}

type LoveMeterImpl struct {
	Value int
}

func NewLoveMeter() LoveMeter {
	return &LoveMeterImpl{Value: 50}
}

func (lm *LoveMeterImpl) Increase(amount int) {
	lm.Value += amount
	if lm.Value > 100 {
		lm.Value = 100
	}
}

func (lm *LoveMeterImpl) Decrease(amount int) {
	lm.Value -= amount
	if lm.Value < 0 {
		lm.Value = 0
	}
}

func (lm *LoveMeterImpl) Get() int {
	return lm.Value
}
