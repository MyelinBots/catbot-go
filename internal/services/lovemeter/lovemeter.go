package lovemeter

import (
	"fmt"
	"strings"
)

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)
	Get(player string) int
	GetLoveBar(player string) string
}

type LoveMeterImpl struct {
	Values map[string]int
}

func NewLoveMeter() LoveMeter {
	return &LoveMeterImpl{
		Values: make(map[string]int),
	}
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

func (lm *LoveMeterImpl) GetLoveBar(player string) string {
	love := lm.Get(player)
	return fmt.Sprintf("%s", lm.generateLoveBar(love))
}

func (lm *LoveMeterImpl) generateLoveBar(percent int) string {
	// Generate a love bar based on the percentage
	// This is a placeholder implementation
	// show percentage as a bar of hearts with a total of 10 hearts being 100%
	return fmt.Sprintf("[%-20s]", strings.Repeat("❤️", percent/10))
}
