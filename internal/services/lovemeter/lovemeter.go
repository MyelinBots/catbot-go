package lovemeter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)
	Get(player string) int
	GetLoveBar(player string) string
	GetMood(player string) string
	StatusLine(player string) string

	// decrease 1 % once a day (only those who have reached 100%)
	DailyDecayAll(ctx context.Context) error
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

// --------------------
// helpers
// --------------------

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func (lm *LoveMeterImpl) setCache(key string, v int) {
	lm.mu.Lock()
	lm.values[key] = v
	lm.mu.Unlock()
}

func (lm *LoveMeterImpl) getCache(key string) (int, bool) {
	lm.mu.RLock()
	v, ok := lm.values[key]
	lm.mu.RUnlock()
	return v, ok
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (lm *LoveMeterImpl) persistLove(key string, love int) {
	// cache values first (field times interaction/decay will be handled in another layer)
	_ = lm.catPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      key,
		LoveMeter: love,
		Network:   lm.Network,
		Channel:   lm.Channel,
	})
}

// --------------------
// core API
// --------------------

func (lm *LoveMeterImpl) Increase(player string, amount int) {
	key := norm(player)
	current := lm.Get(key)

	newVal := clamp(current+amount, 0, 100)
	lm.setCache(key, newVal)
	lm.persistLove(key, newVal)
}

func (lm *LoveMeterImpl) Decrease(player string, amount int) {
	key := norm(player)
	current := lm.Get(key)

	newVal := clamp(current-amount, 0, 100)
	lm.setCache(key, newVal)
	lm.persistLove(key, newVal)
}

func (lm *LoveMeterImpl) Get(player string) int {
	key := norm(player)

	// cache first
	if v, ok := lm.getCache(key); ok {
		return v
	}

	// DB fallback
	fp, err := lm.catPlayerRepo.GetPlayerByName(context.Background(), key, lm.Network, lm.Channel)
	if err != nil || fp == nil {
		lm.setCache(key, 0)
		return 0
	}

	lm.setCache(key, fp.LoveMeter)
	return fp.LoveMeter
}

func (lm *LoveMeterImpl) GetLoveBar(player string) string {
	love := lm.Get(player)
	hearts := clamp(love, 0, 100) / 10 // 0..10
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

func (lm *LoveMeterImpl) StatusLine(player string) string {
	love := lm.Get(player)
	return fmt.Sprintf("%d%% %s %s", love, lm.GetMood(player), lm.GetLoveBar(player))
}

// --------------------
// daily decay (DB-driven)
// --------------------

// DailyDecayAll:
// - fine user who reached 100% (repo.ListPlayersAt100)
// - if today hasn't interacted and hasn't decayed today -> decrease 1% and set LastDecayAt
func (lm *LoveMeterImpl) DailyDecayAll(ctx context.Context) error {
	now := time.Now()

	players, err := lm.catPlayerRepo.ListPlayersAtOrAbove(ctx, lm.Network, lm.Channel, 100)
	if err != nil {
		return err
	}

	for _, p := range players {
		// to prevent double decay in a day
		if p.LastDecayAt != nil && sameDay(*p.LastDecayAt, now) {
			continue
		}
		// if today has interacted, don't decay
		if p.LastInteractedAt != nil && sameDay(*p.LastInteractedAt, now) {
			continue
		}

		// decrease 5 and update cache/DB love_meter
		lm.Decrease(p.Name, 5)

		// set last_decay_at
		_ = lm.catPlayerRepo.SetDecayAt(ctx, p.Name, p.Network, p.Channel, now)
	}

	return nil
}

func (lm *LoveMeterImpl) DailyDecayWithWarning(ctx context.Context) ([]string, error) {
	now := time.Now()

	players, err := lm.catPlayerRepo.ListPlayersAtOrAbove(ctx, lm.Network, lm.Channel, 100)
	if err != nil {
		return nil, err
	}

	var announcements []string

	for _, p := range players {
		// to prevent double decay in a day
		if p.LastDecayAt != nil && sameDay(*p.LastDecayAt, now) {
			continue
		}
		// if today has interacted, don't decay
		if p.LastInteractedAt != nil && sameDay(*p.LastInteractedAt, now) {
			continue
		}

		oldLove := p.LoveMeter
		newLove := oldLove - 5
		if newLove < 0 {
			newLove = 0
		}

		// decrease 5 (will persist)
		lm.Decrease(p.Name, 5)

		// set last_decay_at
		_ = lm.catPlayerRepo.SetDecayAt(ctx, p.Name, p.Network, p.Channel, now)

		// ✅ warning ครั้งแรกเท่านั้น ตอน 100 -> 95
		if oldLove == 100 && !p.PerfectDropWarned {
			// warning message
			announcements = append(announcements,
				fmt.Sprintf("😿 Purrito is waiting but %s did not come today, the perfect bond has begun to fade (100%% → 95%%) 🐾", p.Name),
			)
			_ = lm.catPlayerRepo.SetPerfectDropWarned(ctx, p.Name, p.Network, p.Channel, true)
		}
	}

	return announcements, nil
}
