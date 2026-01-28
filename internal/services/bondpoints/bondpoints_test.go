package bondpoints

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

// mockCatPlayerRepo is a simple in-memory mock for testing
type mockCatPlayerRepo struct {
	mu      sync.RWMutex
	players map[string]*cat_player.CatPlayer
}

func newMockRepo() *mockCatPlayerRepo {
	return &mockCatPlayerRepo{
		players: make(map[string]*cat_player.CatPlayer),
	}
}

func (m *mockCatPlayerRepo) key(name, network, channel string) string {
	return strings.ToLower(name) + "|" + strings.ToLower(network) + "|" + strings.ToLower(channel)
}

func (m *mockCatPlayerRepo) GetPlayerByID(ctx context.Context, id string) (*cat_player.CatPlayer, error) {
	return nil, nil
}

func (m *mockCatPlayerRepo) GetPlayerByName(ctx context.Context, name, network, channel string) (*cat_player.CatPlayer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p := m.players[m.key(name, network, channel)]
	if p != nil {
		// Return a copy
		cp := *p
		return &cp, nil
	}
	return nil, nil
}

func (m *mockCatPlayerRepo) GetAllPlayers(ctx context.Context, network, channel string) ([]*cat_player.CatPlayer, error) {
	return nil, nil
}

func (m *mockCatPlayerRepo) UpsertPlayer(ctx context.Context, player *cat_player.CatPlayer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(player.Name, player.Network, player.Channel)
	m.players[k] = player
	return nil
}

func (m *mockCatPlayerRepo) TopLoveMeter(ctx context.Context, network, channel string, limit int) ([]*cat_player.CatPlayer, error) {
	return nil, nil
}

func (m *mockCatPlayerRepo) TouchInteraction(ctx context.Context, name, network, channel string, t time.Time) error {
	return nil
}

func (m *mockCatPlayerRepo) SetDecayAt(ctx context.Context, name, network, channel string, t time.Time) error {
	return nil
}

func (m *mockCatPlayerRepo) ListPlayersAtOrAbove(ctx context.Context, network, channel string, minLove int) ([]*cat_player.CatPlayer, error) {
	return nil, nil
}

func (m *mockCatPlayerRepo) SetPerfectDropWarned(ctx context.Context, name, network, channel string, warned bool) error {
	return nil
}

func (m *mockCatPlayerRepo) AddBondPoints(ctx context.Context, name, network, channel string, delta int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.BondPoints += delta
	}
	return nil
}

func (m *mockCatPlayerRepo) SetBondPointsAt(ctx context.Context, name, network, channel string, t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.LastBondPointsAt = &t
	}
	return nil
}

func (m *mockCatPlayerRepo) SetBondPointStreak(ctx context.Context, name, network, channel string, streak int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.BondPointStreak = streak
	}
	return nil
}

func (m *mockCatPlayerRepo) SetHighestBondStreak(ctx context.Context, name, network, channel string, streak int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.HighestBondStreak = streak
	}
	return nil
}

func (m *mockCatPlayerRepo) AddGiftsUnlocked(ctx context.Context, name, network, channel string, giftMask int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.GiftsUnlocked |= giftMask
	}
	return nil
}

func (m *mockCatPlayerRepo) SetGiftsUnlocked(ctx context.Context, name, network, channel string, giftsUnlocked int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.GiftsUnlocked = giftsUnlocked
	}
	return nil
}

func (m *mockCatPlayerRepo) SetLoveMeter(ctx context.Context, nick, network, channel string, love int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(nick, network, channel)
	if p, ok := m.players[k]; ok {
		p.LoveMeter = love
	}
	return nil
}

// Tests

func TestPointsForStreak(t *testing.T) {
	tests := []struct {
		streak int
		want   int
	}{
		{1, 2},   // base 2
		{6, 2},   // still 2 (floor(6/7) = 0)
		{7, 3},   // 2 + floor(7/7) = 3
		{13, 3},  // 2 + floor(13/7) = 3
		{14, 4},  // 2 + floor(14/7) = 4
		{21, 5},  // 2 + floor(21/7) = 5
		{28, 6},  // 2 + floor(28/7) = 6
		{35, 7},  // 2 + floor(35/7) = 7 (max)
		{100, 7}, // capped at 7
		{0, 2},   // edge case
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := pointsForStreak(tt.streak)
			if got != tt.want {
				t.Errorf("pointsForStreak(%d) = %d, want %d", tt.streak, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo)
	if svc == nil {
		t.Fatal("New() returned nil")
	}
}

func TestRecordBondedInteraction_NewPlayer(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo)

	ctx := context.Background()
	result, err := svc.RecordBondedInteraction(ctx, "player1", "testnet", "#testchan")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Streak != 1 {
		t.Errorf("expected streak 1, got %d", result.Streak)
	}

	if result.AwardedPoints != 2 {
		t.Errorf("expected 2 points awarded, got %d", result.AwardedPoints)
	}

	if result.TotalPoints != 2 {
		t.Errorf("expected 2 total points, got %d", result.TotalPoints)
	}
}

func TestRecordBondedInteraction_SameDay(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo)
	ctx := context.Background()

	// First interaction
	result1, _ := svc.RecordBondedInteraction(ctx, "player1", "testnet", "#testchan")
	if result1.AwardedPoints == 0 {
		t.Error("first interaction should award points")
	}

	// Second interaction same day
	result2, _ := svc.RecordBondedInteraction(ctx, "player1", "testnet", "#testchan")
	if result2.AwardedPoints != 0 {
		t.Errorf("second interaction same day should award 0 points, got %d", result2.AwardedPoints)
	}

	// Should still report correct totals
	if result2.TotalPoints != result1.TotalPoints {
		t.Errorf("total points should be unchanged, got %d vs %d", result2.TotalPoints, result1.TotalPoints)
	}
}

func TestRecordBondedInteraction_HighestStreak(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo).(*Impl)
	ctx := context.Background()

	// Setup: player with existing highest streak
	now := svc.nyNow()
	yesterday := now.AddDate(0, 0, -1)

	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:              "player1",
		Network:           "testnet",
		Channel:           "#testchan",
		BondPointStreak:   5,
		HighestBondStreak: 10,
		LastBondPointsAt:  &yesterday,
	})

	result, _ := svc.RecordBondedInteraction(ctx, "player1", "testnet", "#testchan")

	// Streak should be 6 (5+1 for consecutive day)
	if result.Streak != 6 {
		t.Errorf("expected streak 6, got %d", result.Streak)
	}

	// Highest streak should still be 10
	if result.HighestStreak != 10 {
		t.Errorf("expected highest streak 10, got %d", result.HighestStreak)
	}
}

func TestRecordBondedInteraction_NewHighestStreak(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo).(*Impl)
	ctx := context.Background()

	// Setup: player about to beat their highest
	now := svc.nyNow()
	yesterday := now.AddDate(0, 0, -1)

	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:              "player1",
		Network:           "testnet",
		Channel:           "#testchan",
		BondPointStreak:   9,
		HighestBondStreak: 9,
		LastBondPointsAt:  &yesterday,
	})

	result, _ := svc.RecordBondedInteraction(ctx, "player1", "testnet", "#testchan")

	// Should have new highest
	if result.HighestStreak != 10 {
		t.Errorf("expected new highest streak 10, got %d", result.HighestStreak)
	}
}

func TestRecordBondedInteraction_StreakReset(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo).(*Impl)
	ctx := context.Background()

	// Setup: player who missed a day
	now := svc.nyNow()
	twoDaysAgo := now.AddDate(0, 0, -2)

	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:              "player1",
		Network:           "testnet",
		Channel:           "#testchan",
		BondPointStreak:   10,
		HighestBondStreak: 10,
		LastBondPointsAt:  &twoDaysAgo,
	})

	result, _ := svc.RecordBondedInteraction(ctx, "player1", "testnet", "#testchan")

	// Streak should reset to 1
	if result.Streak != 1 {
		t.Errorf("expected streak reset to 1, got %d", result.Streak)
	}

	// Highest should remain
	if result.HighestStreak != 10 {
		t.Errorf("highest streak should remain 10, got %d", result.HighestStreak)
	}
}

func TestSameDayNY(t *testing.T) {
	repo := newMockRepo()
	svc := New(repo).(*Impl)

	loc, _ := time.LoadLocation("America/New_York")

	// Same day in NY
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)
	t2 := time.Date(2024, 1, 15, 23, 0, 0, 0, loc)
	if !svc.sameDayNY(t1, t2) {
		t.Error("should be same day")
	}

	// Different days in NY
	t3 := time.Date(2024, 1, 15, 23, 0, 0, 0, loc)
	t4 := time.Date(2024, 1, 16, 1, 0, 0, 0, loc)
	if svc.sameDayNY(t3, t4) {
		t.Error("should be different days")
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		AwardedPoints: 5,
		TotalPoints:   100,
		Streak:        10,
		HighestStreak: 15,
		GiftsUnlocked: 3,
	}

	if r.AwardedPoints != 5 {
		t.Error("AwardedPoints mismatch")
	}
	if r.TotalPoints != 100 {
		t.Error("TotalPoints mismatch")
	}
	if r.Streak != 10 {
		t.Error("Streak mismatch")
	}
	if r.HighestStreak != 15 {
		t.Error("HighestStreak mismatch")
	}
	if r.GiftsUnlocked != 3 {
		t.Error("GiftsUnlocked mismatch")
	}
}
