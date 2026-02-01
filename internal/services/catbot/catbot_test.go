package catbot

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

// mockIRCClient records messages sent
type mockIRCClient struct {
	mu       sync.Mutex
	messages []string
}

func (m *mockIRCClient) Privmsg(channel, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
}

func (m *mockIRCClient) LastMessage() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return ""
	}
	return m.messages[len(m.messages)-1]
}

func (m *mockIRCClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = nil
}

// mockCatPlayerRepo is a simple in-memory mock
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

func TestNewCatBot(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()

	cb := NewCatBot(client, repo, "testnet", "#testchan")

	if cb == nil {
		t.Fatal("NewCatBot returned nil")
	}
	if cb.Channel != "#testchan" {
		t.Errorf("expected channel #testchan, got %s", cb.Channel)
	}
	if cb.Network != "testnet" {
		t.Errorf("expected network testnet, got %s", cb.Network)
	}
}

func TestIsPresent_InitiallyFalse(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	if cb.IsPresent() {
		t.Error("should not be present initially")
	}
}

func TestConsumePresence_WhenNotPresent(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	consumed := cb.ConsumePresence()
	if consumed {
		t.Error("should not consume presence when not present")
	}
}

func TestConsumePresence_WhenPresent(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Simulate presence
	cb.mu.Lock()
	cb.presentUntil = time.Now().Add(5 * time.Minute)
	cb.mu.Unlock()

	consumed := cb.ConsumePresence()
	if !consumed {
		t.Error("should consume presence when present")
	}

	// Should no longer be present after consuming
	if cb.IsPresent() {
		t.Error("should not be present after consuming")
	}
}

func TestAppearTimes(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	now := time.Now()
	later := now.Add(2 * time.Minute)

	cb.mu.Lock()
	cb.lastAppear = now
	cb.nextAppear = later
	cb.mu.Unlock()

	last, next := cb.AppearTimes()
	if !last.Equal(now) {
		t.Error("last appear time mismatch")
	}
	if !next.Equal(later) {
		t.Error("next appear time mismatch")
	}
}

func TestHandleRandomAction_SendsMessage(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	cb.HandleRandomAction(ctx)

	msg := client.LastMessage()
	if !strings.Contains(msg, "meowww") {
		t.Errorf("expected meowww message, got %q", msg)
	}
}

func TestHandleRandomAction_SyncsCatActionsPresence(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()

	// CatActions starts present
	ca := cb.CatActions.(*cat_actions.CatActions)
	if !ca.IsHere() {
		t.Error("CatActions should be here initially")
	}

	cb.HandleRandomAction(ctx)

	// CatActions should still be here after HandleRandomAction
	if !ca.IsHere() {
		t.Error("CatActions should still be here after HandleRandomAction")
	}
}

func TestHandleCatCommand_NoArgs(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "purrito") || !strings.Contains(msg, "help") {
		t.Errorf("expected help message, got %q", msg)
	}
}

func TestHandleCatCommand_InsufficientArgs(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!pet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "purrito") {
		t.Errorf("expected help message, got %q", msg)
	}
}

func TestHandleCatCommand_PetWhenNotHere(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Force Purrito to be absent
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.ForceAbsent()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!pet purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "not here") {
		t.Errorf("expected 'not here' message, got %q", msg)
	}
}

func TestHandleCatCommand_PetWhenHere(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Make Purrito present via CatActions
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!pet purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "love meter") {
		t.Errorf("expected love meter in response, got %q", msg)
	}
}

func TestHandleCatCommand_Status(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Status doesn't require presence
	err := cb.HandleCatCommand(ctx, "!status purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "status") {
		t.Errorf("expected status message, got %q", msg)
	}
}

func TestHandleCatCommand_Catnip(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Make Purrito present first (catnip requires presence)
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(30 * time.Minute)

	err := cb.HandleCatCommand(ctx, "!catnip purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "catnip") && !strings.Contains(msg, "🌿") {
		t.Errorf("expected catnip message, got %q", msg)
	}

	// Purrito should still be here after catnip
	if !ca.IsHere() {
		t.Error("Purrito should still be here after catnip")
	}
}

func TestNormalizeNick(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"player1", "player1"},
		{"Player1", "player1"},
		{"@player1", "player1"},
		{"+player1", "player1"},
		{"~player1", "player1"},
		{"&player1", "player1"},
		{"%player1", "player1"},
		{"~&@%+player1", "player1"},
		{"  @Player1  ", "player1"},
	}

	for _, tt := range tests {
		got := normalizeNick(tt.input)
		if got != tt.want {
			t.Errorf("normalizeNick(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAppendBondProgress_NotBonded(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()

	// Player with <100 love
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 50,
	})

	msg := "test message"
	result := cb.appendBondProgress(ctx, "player1", msg)

	// Should not modify message when not bonded
	if result != msg {
		t.Errorf("expected unchanged message when not bonded, got %q", result)
	}
}

func TestAppendBondProgress_Bonded(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()

	// Player with 100 love (bonded)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})

	msg := "test message"
	result := cb.appendBondProgress(ctx, "player1", msg)

	// Should add bond progress info
	if !strings.Contains(result, "Streak") || !strings.Contains(result, "BondPoints") {
		t.Errorf("expected bond progress info when bonded, got %q", result)
	}
}

func TestPresenceDuration(t *testing.T) {
	if presenceDuration != 3*time.Minute {
		t.Errorf("presenceDuration = %v, want 3m", presenceDuration)
	}
}

func TestHandleCatCommand_Feed(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Make Purrito present via CatActions
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!feed purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "love meter") {
		t.Errorf("expected love meter in feed response, got %q", msg)
	}
}

func TestHandleCatCommand_Love(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!love purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "love meter") {
		t.Errorf("expected love meter in love response, got %q", msg)
	}
}

func TestHandleCatCommand_Laser(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!laser purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "🔦") {
		t.Errorf("expected laser emoji in response, got %q", msg)
	}
}

func TestHandleCatCommand_Slap(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Slap doesn't require presence
	err := cb.HandleCatCommand(ctx, "!slap purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	// First slap should be a warning
	if msg == "" {
		t.Error("expected response for slap")
	}
}

func TestHandleCatCommand_NotPurrito(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!pet someone_else")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	// Should get rejection for non-purrito target
	if msg == "" {
		t.Error("expected response for non-purrito target")
	}
}

func TestHandleCatCommand_UnknownAction(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!unknown purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "tilts its head") {
		t.Errorf("expected confused response for unknown action, got %q", msg)
	}
}

func TestHandleCatCommand_BondedPlayer(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()

	// Setup bonded player
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})

	// Purrito starts present (no need for EnsureHere)

	ctx = context_manager.SetNickContext(ctx, "player1")

	err := cb.HandleCatCommand(ctx, "!pet purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	// Pet has 60% accept / 40% reject due to randomness.
	// On accept: love stays at 100 (capped), bond progress is appended.
	// On reject: love drops to 99, bond progress is skipped.
	if strings.Contains(msg, "not here") {
		t.Errorf("should not get 'not here' when purrito is present, got %q", msg)
	}
	if strings.Contains(msg, "100%") {
		if !strings.Contains(msg, "Streak") || !strings.Contains(msg, "BondPoints") {
			t.Errorf("expected bond progress when love is still 100%%, got %q", msg)
		}
	}
}

func TestCatBotTimes(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	if len(cb.times) == 0 {
		t.Error("times should not be empty")
	}
	if cb.times[0] != 120 {
		t.Errorf("expected first time to be 120 (2 min), got %d", cb.times[0])
	}
}

func TestAppendBondProgress_NilCatActions(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Replace CatActions with nil-like behavior
	cb.CatActions = nil

	ctx := context.Background()
	msg := "test message"

	// This should handle nil gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("appendBondProgress panicked: %v", r)
		}
	}()

	// Since CatActions is nil, type assertion will fail
	result := cb.appendBondProgress(ctx, "player1", msg)
	if result != msg {
		t.Errorf("expected unchanged message with nil CatActions, got %q", result)
	}
}

func TestAppendBondProgress_WithGiftUnlock(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()

	// Setup bonded player with a streak that will unlock a gift
	yesterday := time.Now().AddDate(0, 0, -1)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:              "player1",
		Network:           "testnet",
		Channel:           "#testchan",
		LoveMeter:         100,
		BondPointStreak:   6, // Will become 7, unlocking first gift
		HighestBondStreak: 6,
		LastBondPointsAt:  &yesterday,
	})

	msg := "test message"
	result := cb.appendBondProgress(ctx, "player1", msg)

	// Should include gift unlock
	if !strings.Contains(result, "🎁") {
		t.Errorf("expected gift unlock emoji, got %q", result)
	}
}

func TestHandleCatCommand_WithActionPrefix(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Test with ! prefix
	err := cb.HandleCatCommand(ctx, "!pet purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "love meter") {
		t.Errorf("expected love meter in response, got %q", msg)
	}
}

func TestHandleRandomAction_Multiple(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()

	// Multiple appearances
	cb.HandleRandomAction(ctx)
	cb.HandleRandomAction(ctx)
	cb.HandleRandomAction(ctx)

	// Should have sent multiple messages
	client.mu.Lock()
	count := len(client.messages)
	client.mu.Unlock()

	if count < 3 {
		t.Errorf("expected at least 3 messages, got %d", count)
	}
}

func TestHandleCatCommand_ActionsRequirePresence(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Force Purrito to be absent
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.ForceAbsent()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Pet should fail when not here
	cb.HandleCatCommand(ctx, "!pet purrito")
	msg1 := client.LastMessage()
	if !strings.Contains(msg1, "not here") {
		t.Errorf("expected 'not here' when Purrito absent, got %q", msg1)
	}

	// Catnip should also fail when not here
	client.Clear()
	cb.HandleCatCommand(ctx, "!catnip purrito")
	msg2 := client.LastMessage()
	if !strings.Contains(msg2, "not here") {
		t.Errorf("catnip should require presence, got %q", msg2)
	}

	// Make Purrito present
	ca.EnsureHere(30 * time.Minute)

	// Now pet should work
	client.Clear()
	cb.HandleCatCommand(ctx, "!pet purrito")
	msg3 := client.LastMessage()
	if strings.Contains(msg3, "not here") {
		t.Errorf("expected pet to work when Purrito is here, got %q", msg3)
	}
}

func TestNormalizeNick_EmptyString(t *testing.T) {
	result := normalizeNick("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestNormalizeNick_OnlyPrefixes(t *testing.T) {
	result := normalizeNick("~&@%+")
	if result != "" {
		t.Errorf("expected empty string for only prefixes, got %q", result)
	}
}

func TestStart_QuickCancel(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	// Use a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run Start in a goroutine
	done := make(chan struct{})
	go func() {
		cb.Start(ctx)
		close(done)
	}()

	// Wait for it to finish
	select {
	case <-done:
		// Good - it exited properly
	case <-time.After(2 * time.Second):
		t.Error("Start did not exit when context was cancelled")
	}

	// Should have sent at least one message (the initial appear)
	client.mu.Lock()
	count := len(client.messages)
	client.mu.Unlock()

	if count < 1 {
		t.Errorf("expected at least 1 message from Start, got %d", count)
	}
}

func TestHandleRandomAction_WanderOffMessage(t *testing.T) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := NewCatBot(client, repo, "testnet", "#testchan")

	ctx := context.Background()
	cb.HandleRandomAction(ctx)

	// Initial message
	client.mu.Lock()
	initialCount := len(client.messages)
	client.mu.Unlock()

	if initialCount < 1 {
		t.Error("expected at least 1 message from HandleRandomAction")
	}

	// Wait a bit longer than presence duration - but we can't easily test
	// the wander off because it waits 3 minutes. We just verify the goroutine
	// was spawned without error.
}
