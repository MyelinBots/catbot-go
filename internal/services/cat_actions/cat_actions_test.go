package cat_actions

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
	return p, nil
}

func (m *mockCatPlayerRepo) GetAllPlayers(ctx context.Context, network, channel string) ([]*cat_player.CatPlayer, error) {
	return nil, nil
}

func (m *mockCatPlayerRepo) UpsertPlayer(ctx context.Context, player *cat_player.CatPlayer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(player.Name, player.Network, player.Channel)
	if existing, ok := m.players[k]; ok {
		player.ID = existing.ID
	}
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

func TestNewCatActions(t *testing.T) {
	repo := newMockRepo()
	ca := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)

	if ca == nil {
		t.Fatal("NewCatActions returned nil")
	}

	actions := ca.GetActions()
	if len(actions) == 0 {
		t.Error("GetActions returned empty list")
	}
}

func TestIsHere_InitiallyTrue(t *testing.T) {
	repo := newMockRepo()
	ca := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)

	if !ca.IsHere() {
		t.Error("IsHere() should be true initially")
	}
}

func TestEnsureHere(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Purrito starts present
	if !caImpl.IsHere() {
		t.Error("should be here initially")
	}

	// EnsureHere is a no-op when already present
	caImpl.EnsureHere(5 * time.Minute)
	if !caImpl.IsHere() {
		t.Error("should still be here after EnsureHere while present")
	}

	// Force purrito to be absent (simulate expired presence, no pending spawn)
	caImpl.mu.Lock()
	caImpl.presentUntil = time.Time{}
	caImpl.nextSpawnAt = time.Time{}
	caImpl.mu.Unlock()

	if caImpl.IsHere() {
		t.Error("should not be here after clearing presence")
	}

	// EnsureHere should bring purrito back when absent
	caImpl.EnsureHere(5 * time.Minute)
	if !caImpl.IsHere() {
		t.Error("should be here after EnsureHere when absent")
	}
}

func TestGatePresenceForAction_CatnipRequiresPresence(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Force Purrito to be absent
	caImpl.ForceAbsent()

	// Purrito is NOT here - catnip should be blocked
	ok, msg := caImpl.gatePresenceForAction("catnip")
	if ok {
		t.Error("catnip should be blocked when purrito is not here")
	}
	if !strings.Contains(msg, "not here") {
		t.Errorf("expected 'not here' message, got: %s", msg)
	}

	// Now make Purrito present
	caImpl.EnsureHere(5 * time.Minute)
	ok, _ = caImpl.gatePresenceForAction("catnip")
	if !ok {
		t.Error("catnip should be allowed when purrito is here")
	}
}

func TestGatePresenceForAction_OtherActionsBlocked(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Force Purrito to be absent
	caImpl.ForceAbsent()

	actions := []string{"pet", "love", "feed", "laser"}
	for _, action := range actions {
		ok, msg := caImpl.gatePresenceForAction(action)
		if ok {
			t.Errorf("%s should be blocked when purrito is not here", action)
		}
		if !strings.Contains(msg, "not here") {
			t.Errorf("expected 'not here' message for %s, got: %s", action, msg)
		}
	}
}

func TestGatePresenceForAction_AllowedWhenHere(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	caImpl.EnsureHere(5 * time.Minute)

	actions := []string{"pet", "love", "feed", "laser", "catnip"}
	for _, action := range actions {
		ok, _ := caImpl.gatePresenceForAction(action)
		if !ok {
			t.Errorf("%s should be allowed when purrito is here", action)
		}
	}
}

func TestExecuteAction_NotPurrito(t *testing.T) {
	repo := newMockRepo()
	ca := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)

	result := ca.ExecuteAction("pet", "player1", "someone_else")

	if !strings.Contains(result, "not purrito") && !strings.Contains(result, "confused") &&
		!strings.Contains(result, "Wrong target") && !strings.Contains(result, "tilts") {
		t.Errorf("expected rejection message for non-purrito target, got: %s", result)
	}
}

func TestExecuteAction_PetWhenNotHere(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Force Purrito to be absent
	caImpl.ForceAbsent()

	result := caImpl.ExecuteAction("pet", "player1", "purrito")

	if !strings.Contains(result, "not here") {
		t.Errorf("expected 'not here' message, got: %s", result)
	}
}

func TestExecuteAction_PetWhenHere(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(5 * time.Minute)

	result := caImpl.ExecuteAction("pet", "player1", "purrito")

	// Should get either accept or reject message with love meter info
	if !strings.Contains(result, "love meter") {
		t.Errorf("expected love meter in response, got: %s", result)
	}
}

func TestExecuteAction_LaserWhenHere(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(5 * time.Minute)

	result := caImpl.ExecuteAction("laser", "player1", "purrito")

	// Should contain laser emoji and love meter
	if !strings.Contains(result, "🔦") {
		t.Errorf("expected laser emoji in response, got: %s", result)
	}
	if !strings.Contains(result, "love meter") {
		t.Errorf("expected love meter in response, got: %s", result)
	}
}

func TestExecuteAction_FeedWhenHere(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(5 * time.Minute)

	result := caImpl.ExecuteAction("feed", "player1", "purrito")

	if !strings.Contains(result, "love meter") {
		t.Errorf("expected love meter in response, got: %s", result)
	}
}

func TestExecuteAction_Status(t *testing.T) {
	repo := newMockRepo()
	ca := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)

	// Status doesn't require presence
	result := ca.ExecuteAction("status", "player1", "purrito")

	if !strings.Contains(result, "status") || !strings.Contains(result, "love meter") {
		t.Errorf("expected status message with love meter, got: %s", result)
	}
}

func TestExecuteAction_Catnip(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Force Purrito to be absent
	caImpl.ForceAbsent()

	// Catnip should NOT work when purrito is not here
	result1 := caImpl.ExecuteAction("catnip", "player1", "purrito")
	if !strings.Contains(result1, "not here") {
		t.Errorf("catnip should be blocked when not here, got: %s", result1)
	}

	// Catnip should work when purrito is here
	caImpl.EnsureHere(30 * time.Minute)
	result2 := caImpl.ExecuteAction("catnip", "player2", "purrito")
	if !strings.Contains(result2, "🌿") {
		t.Errorf("expected catnip message with emoji, got: %s", result2)
	}
}

func TestExecuteAction_CatnipCooldown(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(30 * time.Minute)

	// First use should work
	result1 := caImpl.ExecuteAction("catnip", "player1", "purrito")
	if strings.Contains(result1, "already used") {
		t.Error("first catnip use should not be on cooldown")
	}

	// Purrito left after the interaction, force respawn for cooldown test
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// Second use should be on cooldown
	result2 := caImpl.ExecuteAction("catnip", "player1", "purrito")
	if !strings.Contains(result2, "already used") {
		t.Errorf("second catnip use should be on cooldown, got: %s", result2)
	}
}

func TestExecuteAction_SlapWarning(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// First slap - warning only (various warning messages contain different emojis/text)
	result1 := caImpl.ExecuteAction("slap", "player1", "purrito")
	// Check for any of the possible warning indicators
	hasWarningIndicator := strings.Contains(result1, "warning") ||
		strings.Contains(result1, "Warning") ||
		strings.Contains(result1, "⚠️") ||
		strings.Contains(result1, "😾") ||
		strings.Contains(result1, "😿") ||
		strings.Contains(result1, "😼") ||
		strings.Contains(result1, "gentle") ||
		strings.Contains(result1, "shocked")
	if !hasWarningIndicator {
		t.Errorf("first slap should give warning, got: %s", result1)
	}
	// Should NOT contain "decreased" (that's for second slap punishment)
	if strings.Contains(result1, "decreased") {
		t.Errorf("first slap should not decrease love meter, got: %s", result1)
	}

	// Second slap - punishment
	result2 := caImpl.ExecuteAction("slap", "player1", "purrito")
	if !strings.Contains(result2, "love meter") || !strings.Contains(result2, "decreased") {
		t.Errorf("second slap should show love decrease, got: %s", result2)
	}
}

func TestExecuteAction_UnknownAction(t *testing.T) {
	repo := newMockRepo()
	ca := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)

	result := ca.ExecuteAction("unknown_action", "player1", "purrito")

	if !strings.Contains(result, "tilts its head") {
		t.Errorf("expected confused message for unknown action, got: %s", result)
	}
}

func TestCatnipRemaining(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(30 * time.Minute)

	// Initially no cooldown
	if caImpl.CatnipOnCooldown("player1") {
		t.Error("should not be on cooldown initially")
	}

	// Use catnip
	caImpl.ExecuteAction("catnip", "player1", "purrito")

	// Now should be on cooldown
	if !caImpl.CatnipOnCooldown("player1") {
		t.Error("should be on cooldown after use")
	}

	remaining := caImpl.CatnipRemaining("player1")
	if remaining <= 0 || remaining > 24*time.Hour {
		t.Errorf("remaining time should be between 0 and 24h, got: %v", remaining)
	}
}

func TestGetRandomAction(t *testing.T) {
	repo := newMockRepo()
	ca := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute)

	// Run multiple times to check randomness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		action := ca.GetRandomAction()
		if action == "" {
			t.Error("GetRandomAction returned empty string")
		}
		seen[action] = true
	}

	// Should see at least a few different actions
	if len(seen) < 3 {
		t.Errorf("expected more variation in random actions, only saw %d different actions", len(seen))
	}
}

func TestNormalizeAction(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"pet", "pet"},
		{"PET", "pet"},
		{"!pet", "pet"},
		{"!PET", "pet"},
		{"  pet  ", "pet"},
		{"  !PET  ", "pet"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeAction(tt.input)
			if got != tt.want {
				t.Errorf("normalizeAction(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCatnipRequiresPresence(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Force Purrito to be absent
	caImpl.ForceAbsent()

	if caImpl.IsHere() {
		t.Error("should not be here after ForceAbsent")
	}

	// Catnip should NOT work when not here
	result := caImpl.ExecuteAction("catnip", "player1", "purrito")
	if !strings.Contains(result, "not here") {
		t.Errorf("catnip should require presence, got: %s", result)
	}

	// Make purrito present
	caImpl.EnsureHere(5 * time.Minute)

	// Use catnip - purrito should leave after interaction (one interaction per spawn)
	caImpl.ExecuteAction("catnip", "player2", "purrito")

	// Should NOT be here anymore (Purrito leaves after interaction)
	if caImpl.IsHere() {
		t.Error("purrito should leave after catnip interaction")
	}
}

func TestCaseInsensitiveTarget(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(5 * time.Minute)

	targets := []string{"purrito", "PURRITO", "Purrito", "PuRrItO"}
	for _, target := range targets {
		result := caImpl.ExecuteAction("pet", "player1", target)
		if strings.Contains(result, "does not want") {
			t.Errorf("target %q should be recognized as purrito", target)
		}
	}
}

// TestCatnipIndependentCooldowns tests that User A and User B have independent
// 24-hour cooldowns for catnip usage.
func TestCatnipIndependentCooldowns(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Ensure Purrito is present
	caImpl.EnsureHere(30 * time.Minute)

	// User A uses catnip - should succeed
	resultA1 := caImpl.ExecuteAction("catnip", "userA", "purrito")
	if strings.Contains(resultA1, "already used") {
		t.Errorf("User A first catnip should succeed, got: %s", resultA1)
	}
	if !strings.Contains(resultA1, "🌿") {
		t.Errorf("User A catnip response should contain catnip emoji, got: %s", resultA1)
	}

	// Purrito left after interaction, force respawn for cooldown test
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// User A tries again - should be blocked (cooldown), and Purrito leaves
	resultA2 := caImpl.ExecuteAction("catnip", "userA", "purrito")
	if !strings.Contains(resultA2, "already used") {
		t.Errorf("User A second catnip should be blocked, got: %s", resultA2)
	}

	// Purrito left even on cooldown rejection, force respawn for User B
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// User B uses catnip - should succeed (independent cooldown)
	resultB1 := caImpl.ExecuteAction("catnip", "userB", "purrito")
	if strings.Contains(resultB1, "already used") {
		t.Errorf("User B first catnip should succeed (independent from A), got: %s", resultB1)
	}
	if !strings.Contains(resultB1, "🌿") {
		t.Errorf("User B catnip response should contain catnip emoji, got: %s", resultB1)
	}

	// Purrito left after interaction, force respawn for cooldown test
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// User B tries again - should be blocked
	resultB2 := caImpl.ExecuteAction("catnip", "userB", "purrito")
	if !strings.Contains(resultB2, "already used") {
		t.Errorf("User B second catnip should be blocked, got: %s", resultB2)
	}

	// Both users are now on cooldown (24 hours each)
	if !caImpl.CatnipOnCooldown("userA") {
		t.Error("User A should still be on cooldown")
	}
	if !caImpl.CatnipOnCooldown("userB") {
		t.Error("User B should still be on cooldown")
	}

	// User A's cooldown doesn't affect User B's remaining time (and vice versa)
	remainingA := caImpl.CatnipRemaining("userA")
	remainingB := caImpl.CatnipRemaining("userB")

	// Both should have ~24h remaining (User A used first, so should have slightly less)
	if remainingA <= 0 || remainingA > 24*time.Hour {
		t.Errorf("User A remaining should be between 0 and 24h, got: %v", remainingA)
	}
	if remainingB <= 0 || remainingB > 24*time.Hour {
		t.Errorf("User B remaining should be between 0 and 24h, got: %v", remainingB)
	}

	// User A used first, so should have less remaining time than User B
	if remainingA >= remainingB {
		t.Errorf("User A (used first) should have less remaining time than User B, A: %v, B: %v", remainingA, remainingB)
	}
}

// TestCatnipCooldownConsumesPresence tests that when catnip is rejected due to
// cooldown, Purrito still leaves (one interaction per spawn, even failed attempts)
func TestCatnipCooldownConsumesPresence(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)

	// Make Purrito present
	caImpl.EnsureHere(30 * time.Minute)

	// User A uses catnip successfully
	result1 := caImpl.ExecuteAction("catnip", "userA", "purrito")
	if strings.Contains(result1, "already used") {
		t.Error("first catnip should succeed")
	}

	// Purrito should have left after the successful interaction
	if caImpl.IsHere() {
		t.Error("purrito should leave after successful catnip (one interaction per spawn)")
	}

	// Force respawn Purrito to test cooldown rejection
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// User A tries again - should be rejected due to cooldown
	result2 := caImpl.ExecuteAction("catnip", "userA", "purrito")
	if !strings.Contains(result2, "already used") {
		t.Errorf("second catnip should be on cooldown, got: %s", result2)
	}

	// Purrito should have LEFT (cooldown rejection still consumes presence)
	if caImpl.IsHere() {
		t.Error("purrito should leave even after catnip rejection due to cooldown")
	}
}

// TestCatnipCooldownNickNormalization tests that nick prefixes don't create separate cooldowns
func TestCatnipCooldownNickNormalization(t *testing.T) {
	repo := newMockRepo()
	caImpl := NewCatActions(repo, "testnet", "#testchan", 30*time.Minute, 30*time.Minute, 30*time.Minute).(*CatActions)
	caImpl.EnsureHere(30 * time.Minute)

	// User uses catnip as "player1"
	result1 := caImpl.ExecuteAction("catnip", "player1", "purrito")
	if strings.Contains(result1, "already used") {
		t.Error("First catnip should succeed")
	}

	// Purrito left after interaction, force respawn to test cooldown with normalized nicks
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// Same user with @ prefix should be blocked (same cooldown)
	result2 := caImpl.ExecuteAction("catnip", "@player1", "purrito")
	if !strings.Contains(result2, "already used") {
		t.Errorf("@player1 should share cooldown with player1, got: %s", result2)
	}

	// Purrito left even on cooldown rejection, force respawn
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// Same user with different case should be blocked
	result3 := caImpl.ExecuteAction("catnip", "PLAYER1", "purrito")
	if !strings.Contains(result3, "already used") {
		t.Errorf("PLAYER1 should share cooldown with player1, got: %s", result3)
	}

	// Purrito left even on cooldown rejection, force respawn
	caImpl.ForceAbsent()
	caImpl.EnsureHere(30 * time.Minute)

	// Same user with + prefix should be blocked
	result4 := caImpl.ExecuteAction("catnip", "+Player1", "purrito")
	if !strings.Contains(result4, "already used") {
		t.Errorf("+Player1 should share cooldown with player1, got: %s", result4)
	}
}
