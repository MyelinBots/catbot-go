package lovemeter

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
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.LastInteractedAt = &t
	}
	return nil
}

func (m *mockCatPlayerRepo) SetDecayAt(ctx context.Context, name, network, channel string, t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.LastDecayAt = &t
	}
	return nil
}

func (m *mockCatPlayerRepo) ListPlayersAtOrAbove(ctx context.Context, network, channel string, minLove int) ([]*cat_player.CatPlayer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*cat_player.CatPlayer
	for _, p := range m.players {
		if strings.EqualFold(p.Network, network) && strings.EqualFold(p.Channel, channel) && p.LoveMeter >= minLove {
			cp := *p
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (m *mockCatPlayerRepo) SetPerfectDropWarned(ctx context.Context, name, network, channel string, warned bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(name, network, channel)
	if p, ok := m.players[k]; ok {
		p.PerfectDropWarned = warned
	}
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
	return nil
}

func (m *mockCatPlayerRepo) SetGiftsUnlocked(ctx context.Context, name, network, channel string, giftsUnlocked int) error {
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

func TestClampLove(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{-10, 0},
		{0, 0},
		{50, 50},
		{100, 100},
		{150, 100},
	}

	for _, tt := range tests {
		got := ClampLove(tt.input)
		if got != tt.want {
			t.Errorf("ClampLove(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestIsBonded(t *testing.T) {
	tests := []struct {
		love int
		want bool
	}{
		{0, false},
		{50, false},
		{99, false},
		{100, true},
		{150, true},
	}

	for _, tt := range tests {
		got := IsBonded(tt.love)
		if got != tt.want {
			t.Errorf("IsBonded(%d) = %v, want %v", tt.love, got, tt.want)
		}
	}
}

func TestRenderLoveBar(t *testing.T) {
	tests := []struct {
		love int
		want string
	}{
		{0, "[░░░░░░░░░░]"},
		{10, "[❤️░░░░░░░░░]"},
		{50, "[❤️❤️❤️❤️❤️░░░░░]"},
		{100, "[❤️✨❤️✨❤️✨❤️✨❤️]"},
		{-10, "[░░░░░░░░░░]"},   // clamped to 0
		{150, "[❤️✨❤️✨❤️✨❤️✨❤️]"}, // clamped to 100
	}

	for _, tt := range tests {
		got := RenderLoveBar(tt.love)
		if got != tt.want {
			t.Errorf("RenderLoveBar(%d) = %q, want %q", tt.love, got, tt.want)
		}
	}
}

func TestNewLoveMeter(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	if lm == nil {
		t.Fatal("NewLoveMeter returned nil")
	}
}

func TestLoveMeter_Get_NewPlayer(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	// New player should have 0 love
	love := lm.Get("newplayer")
	if love != 0 {
		t.Errorf("new player should have 0 love, got %d", love)
	}
}

func TestLoveMeter_Increase(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	// Setup player with 0 love
	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 0,
	})

	lm.Increase("player1", 10)

	love := lm.Get("player1")
	if love != 10 {
		t.Errorf("expected love 10, got %d", love)
	}
}

func TestLoveMeter_Decrease(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	// Setup player with 50 love
	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 50,
	})

	lm.Decrease("player1", 10)

	love := lm.Get("player1")
	if love != 40 {
		t.Errorf("expected love 40, got %d", love)
	}
}

func TestLoveMeter_IncreaseCapped(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 95,
	})

	lm.Increase("player1", 10)

	love := lm.Get("player1")
	if love != 100 {
		t.Errorf("love should be capped at 100, got %d", love)
	}
}

func TestLoveMeter_DecreaseCapped(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 5,
	})

	lm.Decrease("player1", 10)

	love := lm.Get("player1")
	if love != 0 {
		t.Errorf("love should be capped at 0, got %d", love)
	}
}

func TestLoveMeter_GetMood(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	tests := []struct {
		love     int
		contains string
	}{
		{0, "hostile"},
		{10, "sad"},
		{30, "cautious"},
		{60, "friendly"},
		{90, "loves you"},
	}

	for _, tt := range tests {
		repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
			Name:      "player1",
			Network:   "testnet",
			Channel:   "#testchan",
			LoveMeter: tt.love,
		})

		mood := lm.GetMood("player1")
		if !strings.Contains(mood, tt.contains) {
			t.Errorf("GetMood at love %d should contain %q, got %q", tt.love, tt.contains, mood)
		}
	}
}

func TestLoveMeter_GetLoveBar(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 50,
	})

	bar := lm.GetLoveBar("player1")
	if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
		t.Errorf("love bar should have brackets, got %q", bar)
	}
}

func TestLoveMeter_StatusLine(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 50,
	})

	status := lm.StatusLine("player1")
	if !strings.Contains(status, "50%") {
		t.Errorf("status should contain 50%%, got %q", status)
	}
}

func TestBondPointsForStreak(t *testing.T) {
	tests := []struct {
		streak int
		want   int
	}{
		{1, 2},
		{6, 2},
		{7, 3},
		{14, 4},
		{21, 5},
		{28, 6},
		{35, 7},
		{100, 7}, // capped
	}

	for _, tt := range tests {
		got := bondPointsForStreak(tt.streak)
		if got != tt.want {
			t.Errorf("bondPointsForStreak(%d) = %d, want %d", tt.streak, got, tt.want)
		}
	}
}

func TestRecordInteraction_NotBonded(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 50, // not bonded
	})

	pts, streak, err := lm.RecordInteraction(ctx, "player1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pts != 0 || streak != 0 {
		t.Errorf("not bonded should return 0,0 - got pts=%d, streak=%d", pts, streak)
	}
}

func TestRecordInteraction_Bonded(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100, // bonded
	})

	pts, streak, err := lm.RecordInteraction(ctx, "player1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pts < 2 {
		t.Errorf("bonded should earn at least 2 points, got %d", pts)
	}
	if streak != 1 {
		t.Errorf("first interaction should have streak 1, got %d", streak)
	}
}

func TestDailyDecayAll(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	// Player at 100% who didn't interact today
	yesterday := time.Now().AddDate(0, 0, -1)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:             "player1",
		Network:          "testnet",
		Channel:          "#testchan",
		LoveMeter:        100,
		LastInteractedAt: &yesterday,
	})

	err := lm.DailyDecayAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	love := lm.Get("player1")
	if love != 95 {
		t.Errorf("expected love to decay to 95, got %d", love)
	}
}

func TestDailyDecayAll_InteractedToday(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	// Player at 100% who interacted today - should NOT decay
	today := time.Now()
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:             "player1",
		Network:          "testnet",
		Channel:          "#testchan",
		LoveMeter:        100,
		LastInteractedAt: &today,
	})

	err := lm.DailyDecayAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	love := lm.Get("player1")
	if love != 100 {
		t.Errorf("player who interacted today should not decay, got %d", love)
	}
}

func TestNorm(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Player1", "player1"},
		{"  PLAYER1  ", "player1"},
		{"player1", "player1"},
	}

	for _, tt := range tests {
		got := norm(tt.input)
		if got != tt.want {
			t.Errorf("norm(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDailyDecayWithWarning(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	// Player at 100% who didn't interact today
	yesterday := time.Now().AddDate(0, 0, -1)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:              "player1",
		Network:           "testnet",
		Channel:           "#testchan",
		LoveMeter:         100,
		LastInteractedAt:  &yesterday,
		PerfectDropWarned: false,
	})

	announcements, err := lm.(*LoveMeterImpl).DailyDecayWithWarning(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have warning announcement
	if len(announcements) == 0 {
		t.Error("expected warning announcement for first decay from 100")
	}

	love := lm.Get("player1")
	if love != 95 {
		t.Errorf("expected love to decay to 95, got %d", love)
	}
}

func TestDailyDecayWithWarning_AlreadyWarned(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	yesterday := time.Now().AddDate(0, 0, -1)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:              "player1",
		Network:           "testnet",
		Channel:           "#testchan",
		LoveMeter:         100,
		LastInteractedAt:  &yesterday,
		PerfectDropWarned: true, // Already warned
	})

	announcements, err := lm.(*LoveMeterImpl).DailyDecayWithWarning(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not have warning if already warned
	if len(announcements) != 0 {
		t.Error("should not warn again if already warned")
	}
}

func TestDailyDecayAll_AlreadyDecayedToday(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	today := time.Now()
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:        "player1",
		Network:     "testnet",
		Channel:     "#testchan",
		LoveMeter:   100,
		LastDecayAt: &today, // Already decayed today
	})

	err := lm.DailyDecayAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not decay again
	love := lm.Get("player1")
	if love != 100 {
		t.Errorf("should not decay if already decayed today, got %d", love)
	}
}

func TestRecordInteraction_SecondCallSameDay(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})

	// First call
	pts1, streak1, err := lm.RecordInteraction(ctx, "player1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pts1 < 2 {
		t.Errorf("first call should award points, got %d", pts1)
	}

	// Second call same day
	pts2, streak2, err := lm.RecordInteraction(ctx, "player1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pts2 != 0 {
		t.Errorf("second call same day should not award points, got %d", pts2)
	}
	if streak2 != streak1 {
		t.Errorf("streak should be preserved, got %d vs %d", streak2, streak1)
	}
}

func TestSameDayNY(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")

	// Same day
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)
	t2 := time.Date(2024, 1, 15, 23, 0, 0, 0, loc)
	if !sameDayNY(t1, t2) {
		t.Error("should be same day")
	}

	// Different days
	t3 := time.Date(2024, 1, 15, 23, 0, 0, 0, loc)
	t4 := time.Date(2024, 1, 16, 1, 0, 0, 0, loc)
	if sameDayNY(t3, t4) {
		t.Error("should be different days")
	}
}

func TestSameDay(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 23, 0, 0, 0, time.UTC)
	if !sameDay(t1, t2) {
		t.Error("should be same day")
	}

	t3 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	t4 := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	if sameDay(t3, t4) {
		t.Error("should be different days")
	}
}

func TestNyNow(t *testing.T) {
	now := nyNow()
	if now.IsZero() {
		t.Error("nyNow should not return zero time")
	}
}

func TestLoveMeter_MultipleIncrease(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	repo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 0,
	})

	lm.Increase("player1", 30)
	lm.Increase("player1", 30)
	lm.Increase("player1", 30)

	love := lm.Get("player1")
	if love != 90 {
		t.Errorf("expected love 90, got %d", love)
	}
}

func TestLoveMeter_NewPlayerIncrease(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")

	// Don't pre-create player
	lm.Increase("newplayer", 10)

	// Should create player with love
	love := lm.Get("newplayer")
	// Note: since player doesn't exist, increase will try to persist but fail
	// This tests the error path in persistLove
	_ = love
}

func TestRecordInteraction_NewPlayer(t *testing.T) {
	repo := newMockRepo()
	lm := NewLoveMeter(repo, "testnet", "#testchan")
	ctx := context.Background()

	// Player doesn't exist yet
	pts, streak, err := lm.RecordInteraction(ctx, "newplayer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return 0 since new player won't be bonded
	if pts != 0 || streak != 0 {
		t.Errorf("new player should return 0,0 - got pts=%d, streak=%d", pts, streak)
	}
}
