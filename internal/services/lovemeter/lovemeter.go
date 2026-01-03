package lovemeter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

// --------------------
// interface + constructor
// --------------------

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)

	Get(player string) int
	GetLoveBar(player string) string
	GetMood(player string) string
	StatusLine(player string) string

	// decrease once a day (only those who have reached 100%)
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
// normalization + cache
// --------------------

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

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

// --------------------
// love rules (cap + bonded bar)
// --------------------

func ClampLove(love int) int {
	if love < 0 {
		return 0
	}
	if love > 100 {
		return 100
	}
	return love
}

func IsBonded(love int) bool {
	// since love is capped at 100, >= is fine; you can switch to == if you prefer
	return love >= 100
}

func RenderLoveBar(love int) string {
	love = ClampLove(love)

	if IsBonded(love) {
		return "[❤️✨❤️✨❤️✨❤️✨❤️]"
	}

	filled := love / 10 // 0..10
	bar := "["
	for i := 0; i < filled; i++ {
		bar += "❤️"
	}
	for i := filled; i < 10; i++ {
		bar += "░"
	}
	bar += "]"
	return bar
}

// --------------------
// time helpers
// --------------------

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// --------------------
// persistence
// --------------------

func (lm *LoveMeterImpl) persistLove(key string, love int) {
	_ = lm.catPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      key,
		LoveMeter: love,
		Network:   lm.Network,
		Channel:   lm.Channel,
	})
}

// --------------------
// core API (mutations + reads)
// --------------------

func (lm *LoveMeterImpl) Increase(player string, amount int) {
	key := norm(player)
	current := lm.Get(key)

	newVal := ClampLove(current + amount)
	lm.setCache(key, newVal)
	lm.persistLove(key, newVal)
}

func (lm *LoveMeterImpl) Decrease(player string, amount int) {
	key := norm(player)
	current := lm.Get(key)

	newVal := ClampLove(current - amount)
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

	lm.setCache(key, ClampLove(fp.LoveMeter))
	return ClampLove(fp.LoveMeter)
}

func (lm *LoveMeterImpl) GetLoveBar(player string) string {
	return RenderLoveBar(lm.Get(player))
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
// - find users who reached 100% (repo.ListPlayersAtOrAbove(..., 100))
// - if today hasn't interacted and hasn't decayed today -> decrease 5 and set LastDecayAt
func (lm *LoveMeterImpl) DailyDecayAll(ctx context.Context) error {
	now := time.Now()

	players, err := lm.catPlayerRepo.ListPlayersAtOrAbove(ctx, lm.Network, lm.Channel, 100)
	if err != nil {
		return err
	}

	for _, p := range players {
		// prevent double decay in a day
		if p.LastDecayAt != nil && sameDay(*p.LastDecayAt, now) {
			continue
		}
		// if today has interacted, don't decay
		if p.LastInteractedAt != nil && sameDay(*p.LastInteractedAt, now) {
			continue
		}

		lm.Decrease(p.Name, 5)
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
		// prevent double decay in a day
		if p.LastDecayAt != nil && sameDay(*p.LastDecayAt, now) {
			continue
		}
		// if today has interacted, don't decay
		if p.LastInteractedAt != nil && sameDay(*p.LastInteractedAt, now) {
			continue
		}

		oldLove := p.LoveMeter

		lm.Decrease(p.Name, 5)
		_ = lm.catPlayerRepo.SetDecayAt(ctx, p.Name, p.Network, p.Channel, now)

		// warning only once: 100 -> 95
		if oldLove == 100 && !p.PerfectDropWarned {
			announcements = append(announcements,
				fmt.Sprintf("😿 Purrito is waiting but %s did not come today, the perfect bond has begun to fade (100%% → 95%%) 🐾", p.Name),
			)
			_ = lm.catPlayerRepo.SetPerfectDropWarned(ctx, p.Name, p.Network, p.Channel, true)
		}
	}

	return announcements, nil
}
