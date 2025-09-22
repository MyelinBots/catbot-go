package lovemeter

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)
	Get(player string) int
	GetLoveBar(player string) string
	GetMood(player string) string
	// Convenience: "42% 😺 friendly [❤️❤️❤️❤️░░░░░░]"
	StatusLine(player string) string
}

type LoveMeterImpl struct {
	mu            sync.RWMutex
	values        map[string]int
	catPlayerRepo cat_player.CatPlayerRepository
	Network       string
	Channel       string
}

func NewLoveMeter(catPlayerRepo cat_player.CatPlayerRepository, network, channel string) LoveMeter {
	return &LoveMeterImpl{
		values:        make(map[string]int),
		catPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,
	}
}

// normalize nicknames so DB + cache are consistent
func norm(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (lm *LoveMeterImpl) Increase(player string, amount int) {
	key := norm(player)

	// ensure we start from DB value if not cached
	current := lm.Get(key)
	newVal := current + amount
	if newVal > 100 {
		newVal = 100
	}

	lm.mu.Lock()
	lm.values[key] = newVal
	lm.mu.Unlock()

	// persist
	if err := lm.catPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      key,
		LoveMeter: newVal,
		Network:   lm.Network,
		Channel:   lm.Channel,
	}); err != nil {
		fmt.Printf("Error updating love meter for %s: %v\n", key, err)
	}
}

func (lm *LoveMeterImpl) Decrease(player string, amount int) {
	key := norm(player)

	current := lm.Get(key)
	newVal := current - amount
	if newVal < 0 {
		newVal = 0
	}

	lm.mu.Lock()
	lm.values[key] = newVal
	lm.mu.Unlock()

	if err := lm.catPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      key,
		LoveMeter: newVal,
		Network:   lm.Network,
		Channel:   lm.Channel,
	}); err != nil {
		fmt.Printf("Error updating love meter for %s: %v\n", key, err)
	}
}

func (lm *LoveMeterImpl) Get(player string) int {
	key := norm(player)

	// cache
	lm.mu.RLock()
	if v, ok := lm.values[key]; ok {
		lm.mu.RUnlock()
		return v
	}
	lm.mu.RUnlock()

	// miss → DB
	fp, err := lm.catPlayerRepo.GetPlayerByName(key, lm.Network, lm.Channel)
	if err != nil || fp == nil {
		lm.mu.Lock()
		lm.values[key] = 0
		lm.mu.Unlock()
		return 0
	}

	lm.mu.Lock()
	lm.values[key] = fp.LoveMeter
	lm.mu.Unlock()
	return fp.LoveMeter
}

func (lm *LoveMeterImpl) GetLoveBar(player string) string {
	love := lm.Get(player)
	return lm.generateLoveBar(love)
}

func (lm *LoveMeterImpl) generateLoveBar(percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	hearts := percent / 10 // 0..10
	return fmt.Sprintf("[%s%s]", strings.Repeat("❤️", hearts), strings.Repeat("░", 10-hearts))
}

func (lm *LoveMeterImpl) GetMood(player string) string {
	love := lm.Get(player)
	switch {
	case love == 0:
		return "hostile 😾"
	case love < 20:
		return "sad 😿"
	case love < 50:
		return "cautious 😼"
	case love < 80:
		return "friendly 😺"
	default:
		return "loves you 😻"
	}
}

// StatusLine returns a compact, ready-to-print string like:
// "42% 😺 friendly [❤️❤️❤️❤️░░░░░░]"
func (lm *LoveMeterImpl) StatusLine(player string) string {
	love := lm.Get(player)
	return fmt.Sprintf("%d%% %s %s", love, lm.GetMood(player), lm.GetLoveBar(player))
}
