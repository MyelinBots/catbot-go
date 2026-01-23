package lovemeter

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

// --------------------------------------------------
// Interface + Constructor
// --------------------------------------------------

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)

	Get(player string) int
	GetLoveBar(player string) string
	GetMood(player string) string
	StatusLine(player string) string

	// Returns: (bondPointsAwardedToday, newStreak)
	RecordInteraction(ctx context.Context, player string) (awardedBondPoints int, newStreak int, err error)

	// decrease once a day (only those who have reached 100%)
	DailyDecayAll(ctx context.Context) error
}

type LoveMeterImpl struct {
	catPlayerRepo cat_player.CatPlayerRepository
	Network       string
	Channel       string
}

func NewLoveMeter(catPlayerRepo cat_player.CatPlayerRepository, network, channel string) LoveMeter {
	return &LoveMeterImpl{
		catPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,
	}
}

// --------------------------------------------------
// Normalization
// --------------------------------------------------

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// --------------------------------------------------
// Love Rules (cap + bonded bar)
// --------------------------------------------------

func ClampLove(love int) int {
	if love < 0 {
		return 0
	}
	if love > 100 {
		return 100
	}
	return love
}

func IsBonded(love int) bool { return love >= 100 }

func RenderLoveBar(love int) string {
	love = ClampLove(love)

	if IsBonded(love) {
		return "[❤️✨❤️✨❤️✨❤️✨❤️]"
	}

	filled := love / 10
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

// --------------------------------------------------
// Time Helpers
// --------------------------------------------------

func nyNow() time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func sameDayNY(a, b time.Time) bool {
	loc, _ := time.LoadLocation("America/New_York")
	aa := a.In(loc)
	bb := b.In(loc)
	return aa.Year() == bb.Year() && aa.YearDay() == bb.YearDay()
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// --------------------------------------------------
// Persistence
// --------------------------------------------------

func (lm *LoveMeterImpl) persistLove(key string, love int) {
	err := lm.catPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      key,
		LoveMeter: love,
		Network:   lm.Network,
		Channel:   lm.Channel,
	})
	if err != nil {
		log.Printf("failed to persist love for %s: %v", key, err)
	}
}

// --------------------------------------------------
// Core API (mutations + reads)
// --------------------------------------------------

func (lm *LoveMeterImpl) Increase(player string, amount int) {
	key := norm(player)
	current := lm.Get(key)

	newVal := ClampLove(current + amount)
	lm.persistLove(key, newVal)
}

func (lm *LoveMeterImpl) Decrease(player string, amount int) {
	key := norm(player)
	current := lm.Get(key)

	newVal := ClampLove(current - amount)
	lm.persistLove(key, newVal)
}

func (lm *LoveMeterImpl) Get(player string) int {
	key := norm(player)

	fp, err := lm.catPlayerRepo.GetPlayerByName(context.Background(), key, lm.Network, lm.Channel)
	if err != nil || fp == nil {
		return 0
	}

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

// --------------------------------------------------
// BondPoints (Top Love)
// --------------------------------------------------

// Base: +2
// Bonus: +min(5, floor(streak/7))
// => 2..7 per day
func bondPointsForStreak(streak int) int {
	bonus := int(math.Floor(float64(streak) / 7.0))
	if bonus > 5 {
		bonus = 5
	}
	if bonus < 0 {
		bonus = 0
	}
	return 2 + bonus
}

// RecordInteraction (fixed):
// - TouchInteraction always (for decay)
// - Only when love==100 (bonded)
// - Only once per NY day using LastBondPointsAt
// - Uses CatPlayer.BondPointStreak + repo.SetBondPointStreak
func (lm *LoveMeterImpl) RecordInteraction(ctx context.Context, player string) (awardedBondPoints int, newStreak int, err error) {
	key := norm(player)
	now := nyNow()

	// Always mark interaction time (supports decay logic)
	_ = lm.catPlayerRepo.TouchInteraction(ctx, key, lm.Network, lm.Channel, now)

	// Must load player from DB
	p, err := lm.catPlayerRepo.GetPlayerByName(ctx, key, lm.Network, lm.Channel)
	if err != nil {
		return 0, 0, err
	}
	if p == nil {
		// ensure row exists
		if err := lm.catPlayerRepo.UpsertPlayer(ctx, &cat_player.CatPlayer{
			Name:    key,
			Network: lm.Network,
			Channel: lm.Channel,
		}); err != nil {
			return 0, 0, err
		}
		p, err = lm.catPlayerRepo.GetPlayerByName(ctx, key, lm.Network, lm.Channel)
		if err != nil || p == nil {
			return 0, 0, fmt.Errorf("failed to load player %s", key)
		}
	}

	// gate: bonded only
	if ClampLove(p.LoveMeter) != 100 {
		return 0, 0, nil
	}

	// once per NY day
	if p.LastBondPointsAt != nil && sameDayNY(*p.LastBondPointsAt, now) {
		return 0, p.BondPointStreak, nil
	}

	// streak rule:
	// if last award was yesterday -> streak++, else reset to 1
	newStreak = 1
	if p.LastBondPointsAt != nil {
		yesterday := now.AddDate(0, 0, -1)
		if sameDayNY(*p.LastBondPointsAt, yesterday) {
			newStreak = p.BondPointStreak + 1
		}
	}

	awardedBondPoints = bondPointsForStreak(newStreak)

	// Persist progress
	if err := lm.catPlayerRepo.SetBondPointStreak(ctx, key, lm.Network, lm.Channel, newStreak); err != nil {
		return 0, 0, err
	}
	if err := lm.catPlayerRepo.AddBondPoints(ctx, key, lm.Network, lm.Channel, awardedBondPoints); err != nil {
		return 0, 0, err
	}
	if err := lm.catPlayerRepo.SetBondPointsAt(ctx, key, lm.Network, lm.Channel, now); err != nil {
		return 0, 0, err
	}

	return awardedBondPoints, newStreak, nil
}

// --------------------------------------------------
// Daily Decay (DB-driven)
// --------------------------------------------------

func (lm *LoveMeterImpl) DailyDecayAll(ctx context.Context) error {
	now := time.Now()

	players, err := lm.catPlayerRepo.ListPlayersAtOrAbove(ctx, lm.Network, lm.Channel, 100)
	if err != nil {
		return err
	}

	for _, p := range players {
		if p.LastDecayAt != nil && sameDay(*p.LastDecayAt, now) {
			continue
		}
		if p.LastInteractedAt != nil && sameDay(*p.LastInteractedAt, now) {
			continue
		}

		// decay 100 -> 95
		lm.Decrease(p.Name, 5)
		if err := lm.catPlayerRepo.SetDecayAt(ctx, p.Name, p.Network, p.Channel, now); err != nil {
			log.Printf("failed to set decay at for %s: %v", p.Name, err)
		}

		// reset bond streak on decay
		if err := lm.catPlayerRepo.SetBondPointStreak(ctx, p.Name, p.Network, p.Channel, 0); err != nil {
			log.Printf("failed to reset bond streak for %s: %v", p.Name, err)
		}
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
		if err := lm.catPlayerRepo.SetDecayAt(ctx, p.Name, p.Network, p.Channel, now); err != nil {
			log.Printf("failed to set decay at for %s: %v", p.Name, err)
		}

		// reset bond streak on decay
		if err := lm.catPlayerRepo.SetBondPointStreak(ctx, p.Name, p.Network, p.Channel, 0); err != nil {
			log.Printf("failed to reset bond streak for %s: %v", p.Name, err)
		}

		// warning only once: 100 -> 95
		if oldLove == 100 && !p.PerfectDropWarned {
			announcements = append(announcements,
				fmt.Sprintf("😿 Purrito is waiting but %s did not come today, the perfect bond has begun to fade (100%% → 95%%) 🐾", p.Name),
			)
			if err := lm.catPlayerRepo.SetPerfectDropWarned(ctx, p.Name, p.Network, p.Channel, true); err != nil {
				log.Printf("failed to set perfect drop warned for %s: %v", p.Name, err)
			}
		}
	}

	return announcements, nil
}
