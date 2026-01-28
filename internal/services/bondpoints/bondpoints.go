package bondpoints

import (
	"context"
	"math"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

type Result struct {
	AwardedPoints int // 0 if not awarded today
	TotalPoints   int
	Streak        int
	HighestStreak int
	GiftsUnlocked int
}

type Service interface {
	// Only call when LoveMeter == 100 (bonded)
	RecordBondedInteraction(ctx context.Context, nick, network, channel string) (Result, error)
}

type Impl struct {
	repo cat_player.CatPlayerRepository
	loc  *time.Location
}

func New(repo cat_player.CatPlayerRepository) Service {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}
	return &Impl{repo: repo, loc: loc}
}

func (s *Impl) nyNow() time.Time {
	return time.Now().In(s.loc)
}

func (s *Impl) sameDayNY(a, b time.Time) bool {
	aa := a.In(s.loc)
	bb := b.In(s.loc)
	return aa.Year() == bb.Year() && aa.YearDay() == bb.YearDay()
}

func pointsForStreak(streak int) int {
	bonus := int(math.Floor(float64(streak) / 7.0))
	if bonus > 5 {
		bonus = 5
	}
	if bonus < 0 {
		bonus = 0
	}
	return 2 + bonus // 2..7
}

func (s *Impl) RecordBondedInteraction(ctx context.Context, nick, network, channel string) (Result, error) {
	now := s.nyNow()

	p, err := s.repo.GetPlayerByName(ctx, nick, network, channel)
	if err != nil {
		return Result{}, err
	}
	if p == nil {
		_ = s.repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
			Name:    nick,
			Network: network,
			Channel: channel,
		})
		p, err = s.repo.GetPlayerByName(ctx, nick, network, channel)
		if err != nil || p == nil {
			return Result{}, err
		}
	}

	if p.LastBondPointsAt != nil && s.sameDayNY(*p.LastBondPointsAt, now) {
		return Result{
			AwardedPoints: 0,
			TotalPoints:   p.BondPoints,
			Streak:        p.BondPointStreak,
			HighestStreak: p.HighestBondStreak,
			GiftsUnlocked: p.GiftsUnlocked,
		}, nil
	}

	newStreak := 1
	if p.LastBondPointsAt != nil {
		yesterday := now.AddDate(0, 0, -1)
		if s.sameDayNY(*p.LastBondPointsAt, yesterday) {
			newStreak = p.BondPointStreak + 1
		}
	}

	pts := pointsForStreak(newStreak)

	newHighest := p.HighestBondStreak
	if newStreak > newHighest {
		newHighest = newStreak
	}

	if err := s.repo.SetBondPointStreak(ctx, nick, network, channel, newStreak); err != nil {
		return Result{}, err
	}
	if err := s.repo.AddBondPoints(ctx, nick, network, channel, pts); err != nil {
		return Result{}, err
	}
	if err := s.repo.SetBondPointsAt(ctx, nick, network, channel, now); err != nil {
		return Result{}, err
	}
	if newHighest != p.HighestBondStreak {
		if err := s.repo.SetHighestBondStreak(ctx, nick, network, channel, newHighest); err != nil {
			return Result{}, err
		}
	}

	p2, _ := s.repo.GetPlayerByName(ctx, nick, network, channel)
	total := p.BondPoints + pts
	gifts := p.GiftsUnlocked
	if p2 != nil {
		total = p2.BondPoints
		gifts = p2.GiftsUnlocked
	}

	return Result{
		AwardedPoints: pts,
		TotalPoints:   total,
		Streak:        newStreak,
		HighestStreak: newHighest,
		GiftsUnlocked: gifts,
	}, nil
}
