package commands

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/catbot"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
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
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*cat_player.CatPlayer
	for _, p := range m.players {
		if strings.EqualFold(p.Network, network) && strings.EqualFold(p.Channel, channel) {
			cp := *p
			result = append(result, &cp)
		}
	}

	// Sort by love meter (descending) - simple bubble sort for test
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].LoveMeter > result[i].LoveMeter {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
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

// Helper to create a test setup
func setupTest() (*mockIRCClient, *mockCatPlayerRepo, *catbot.CatBot, CommandController) {
	client := &mockIRCClient{}
	repo := newMockRepo()
	cb := catbot.NewCatBot(client, repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)
	cc := NewCommandController(cb)
	return client, repo, cb, cc
}

// Tests

func TestNewCommandController(t *testing.T) {
	_, _, cb, cc := setupTest()
	if cc == nil {
		t.Fatal("NewCommandController returned nil")
	}
	_ = cb
}

func TestAddCommand(t *testing.T) {
	_, _, _, cc := setupTest()

	called := false
	cc.AddCommand("!test", func(ctx context.Context, message string) error {
		called = true
		return nil
	})

	// Create a mock IRC line
	line := &irc.Line{
		Nick: "player1",
		Args: []string{"#testchan", "!test"},
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("command handler was not called")
	}
}

func TestHandleCommand_NoArgs(t *testing.T) {
	_, _, _, cc := setupTest()

	line := &irc.Line{
		Nick: "player1",
		Args: []string{}, // no args
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should just return nil without error
}

func TestHandleCommand_SingleArg(t *testing.T) {
	_, _, _, cc := setupTest()

	line := &irc.Line{
		Nick: "player1",
		Args: []string{"#testchan"}, // only channel, no message
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleCommand_UnknownCommand(t *testing.T) {
	_, _, _, cc := setupTest()

	line := &irc.Line{
		Nick: "player1",
		Args: []string{"#testchan", "!unknown"},
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Unknown commands should be silently ignored
}

func TestPurritoLaserHandler_NotPurrito(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Target is not purrito
	err := handler(ctx, "!laser someone_else")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not send any message
	if client.LastMessage() != "" {
		t.Errorf("should not send message for non-purrito target")
	}
}

func TestPurritoLaserHandler_WhenNotHere(t *testing.T) {
	client, _, cb, cc := setupTest()

	// Force Purrito to be absent
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.ForceAbsent()

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := handler(ctx, "!laser purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "not here") {
		t.Errorf("expected 'not here' message, got %q", msg)
	}
}

func TestPurritoLaserHandler_WhenHere(t *testing.T) {
	client, _, cb, cc := setupTest()

	// Make Purrito present
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := handler(ctx, "!laser purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "🔦") {
		t.Errorf("expected laser message, got %q", msg)
	}
	if !strings.Contains(msg, "love meter") {
		t.Errorf("expected love meter in response, got %q", msg)
	}
}

func TestPurritoLaserHandler_InsufficientArgs(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := handler(ctx, "!laser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not send any message
	if client.LastMessage() != "" {
		t.Errorf("should not send message for insufficient args")
	}
}

func TestPurritoLaserHandler_CaseInsensitive(t *testing.T) {
	client, _, cb, cc := setupTest()

	// Make Purrito present
	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Test various case combinations
	testCases := []string{
		"!laser PURRITO",
		"!LASER purrito",
		"!Laser Purrito",
	}

	for _, tc := range testCases {
		client.Clear()
		err := handler(ctx, tc)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc, err)
		}

		msg := client.LastMessage()
		if msg == "" {
			t.Errorf("expected response for %q, got empty", tc)
		}
	}
}

func TestAppendBondProgress_CatnipCooldownMessage(t *testing.T) {
	_, repo, _, cc := setupTest()

	ctx := context.Background()

	// Player at 100 love
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})

	ccImpl := cc.(*CommandControllerImpl)

	// Message containing "already used catnip today" should not be modified
	msg := "aww player1, you already used catnip today. Try again in 5 h 30 m."
	result := ccImpl.appendBondProgress(ctx, "player1", msg)

	if result != msg {
		t.Errorf("catnip cooldown message should not be modified, got %q", result)
	}
}

func TestAppendBondProgress_NotBonded(t *testing.T) {
	_, repo, _, cc := setupTest()

	ctx := context.Background()

	// Player with <100 love
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 50,
	})

	ccImpl := cc.(*CommandControllerImpl)

	msg := "test message"
	result := ccImpl.appendBondProgress(ctx, "player1", msg)

	// Should not modify message when not bonded
	if result != msg {
		t.Errorf("expected unchanged message when not bonded, got %q", result)
	}
}

func TestAppendBondProgress_Bonded(t *testing.T) {
	_, repo, _, cc := setupTest()

	ctx := context.Background()

	// Player with 100 love (bonded)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})

	ccImpl := cc.(*CommandControllerImpl)

	msg := "test message"
	result := ccImpl.appendBondProgress(ctx, "player1", msg)

	// Should add bond progress info
	if !strings.Contains(result, "BondPoints") {
		t.Errorf("expected BondPoints info when bonded, got %q", result)
	}
}

func TestHandleCommand_SetsNickContext(t *testing.T) {
	_, _, _, cc := setupTest()

	var capturedNick string
	cc.AddCommand("!testnick", func(ctx context.Context, message string) error {
		capturedNick = context_manager.GetNickContext(ctx)
		return nil
	})

	line := &irc.Line{
		Nick: "testplayer",
		Args: []string{"#testchan", "!testnick"},
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedNick != "testplayer" {
		t.Errorf("expected nick testplayer, got %q", capturedNick)
	}
}

func TestPurritoLaserHandler_NotLaserCommand(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	// Not a laser command
	err := handler(ctx, "!pet purrito")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not send any message
	if client.LastMessage() != "" {
		t.Errorf("should not send message for non-laser command")
	}
}

func TestAppendBondProgress_ErrorFromLoveMeter(t *testing.T) {
	_, repo, _, cc := setupTest()

	ctx := context.Background()

	// Player with 100 love (bonded)
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})

	ccImpl := cc.(*CommandControllerImpl)

	// First call awards points
	msg1 := "first message"
	result1 := ccImpl.appendBondProgress(ctx, "player1", msg1)
	if !strings.Contains(result1, "BondPoints") {
		t.Errorf("first call should include BondPoints, got %q", result1)
	}

	// Second call same day
	msg2 := "second message"
	result2 := ccImpl.appendBondProgress(ctx, "player1", msg2)
	if !strings.Contains(result2, "already earned") {
		t.Errorf("second call should show already earned, got %q", result2)
	}
}

func TestAddCommand_Overwrite(t *testing.T) {
	_, _, _, cc := setupTest()

	called1 := false
	called2 := false

	cc.AddCommand("!test", func(ctx context.Context, message string) error {
		called1 = true
		return nil
	})

	cc.AddCommand("!test", func(ctx context.Context, message string) error {
		called2 = true
		return nil
	})

	line := &irc.Line{
		Nick: "player1",
		Args: []string{"#testchan", "!test"},
	}

	ctx := context.Background()
	cc.HandleCommand(ctx, line)

	if called1 {
		t.Error("first handler should not be called")
	}
	if !called2 {
		t.Error("second handler should be called")
	}
}

func TestHandleCommand_EmptyMessage(t *testing.T) {
	_, _, _, cc := setupTest()

	line := &irc.Line{
		Nick: "player1",
		Args: []string{"#testchan", ""},
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleCommand_WhitespaceMessage(t *testing.T) {
	_, _, _, cc := setupTest()

	line := &irc.Line{
		Nick: "player1",
		Args: []string{"#testchan", "   "},
	}

	ctx := context.Background()
	err := cc.HandleCommand(ctx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPurritoLaserHandler_EmptyMessage(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := handler(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not send any message
	if client.LastMessage() != "" {
		t.Error("should not send message for empty input")
	}
}

func TestPurritoLaserHandler_WithExtraSpaces(t *testing.T) {
	client, _, cb, cc := setupTest()

	ca := cb.CatActions.(*cat_actions.CatActions)
	ca.EnsureHere(5 * time.Minute)

	handler := cc.(*CommandControllerImpl).PurritoLaserHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := handler(ctx, "  !laser   purrito  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if msg == "" {
		t.Error("expected response for laser with extra spaces")
	}
}

func TestCommandController_Interface(t *testing.T) {
	_, _, _, cc := setupTest()

	// Verify cc implements CommandController interface
	var _ CommandController = cc
}

func TestAppendBondProgress_NilLoveMeter(t *testing.T) {
	_, _, _, cc := setupTest()

	ctx := context.Background()
	ccImpl := cc.(*CommandControllerImpl)

	// This tests the nil LoveMeter path
	ccImpl.game.CatActions = nil

	msg := "test message"
	result := ccImpl.appendBondProgress(ctx, "player1", msg)

	// Should return unchanged message
	if result != msg {
		t.Errorf("expected unchanged message with nil CatActions, got %q", result)
	}
}

func TestPurritoHandler(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).PurritoHandler()

	ctx := context.Background()
	ctx = context_manager.SetNickContext(ctx, "player1")

	err := handler(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have sent multiple messages (the help text lines)
	client.mu.Lock()
	count := len(client.messages)
	client.mu.Unlock()

	if count < 5 {
		t.Errorf("expected multiple help messages, got %d", count)
	}
}

func TestTopLove10Handler_NoArgs(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).TopLove10Handler()

	ctx := context.Background()

	err := handler(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not send any message with no args
	if client.LastMessage() != "" {
		t.Error("should not send message with no args")
	}
}

func TestTopLove10Handler_WrongCommand(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).TopLove10Handler()

	ctx := context.Background()

	err := handler(ctx, "!other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not send any message for wrong command
	if client.LastMessage() != "" {
		t.Error("should not send message for wrong command")
	}
}

func TestTopLove10Handler_NoPlayers(t *testing.T) {
	client, _, _, cc := setupTest()

	handler := cc.(*CommandControllerImpl).TopLove10Handler()

	ctx := context.Background()

	err := handler(ctx, "!toplove")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "No love yet") {
		t.Errorf("expected 'No love yet' message, got %q", msg)
	}
}

func TestTopLove10Handler_WithPlayers(t *testing.T) {
	client, repo, _, cc := setupTest()

	ctx := context.Background()

	// Add some players
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player1",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 100,
	})
	repo.UpsertPlayer(ctx, &cat_player.CatPlayer{
		Name:      "player2",
		Network:   "testnet",
		Channel:   "#testchan",
		LoveMeter: 80,
	})

	handler := cc.(*CommandControllerImpl).TopLove10Handler()

	err := handler(ctx, "!toplove")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := client.LastMessage()
	if !strings.Contains(msg, "Top 10") {
		t.Errorf("expected Top 10 message, got %q", msg)
	}
}
