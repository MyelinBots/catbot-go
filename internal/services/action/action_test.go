package action

import (
	"strings"
	"testing"
)

func TestNewCatActions(t *testing.T) {
	ca := NewCatActions()

	if ca == nil {
		t.Fatal("NewCatActions returned nil")
	}

	if ca.LoveMeter != 0 {
		t.Errorf("expected initial love meter 0, got %d", ca.LoveMeter)
	}

	if _, exists := ca.Actions["pet"]; !exists {
		t.Error("expected 'pet' action to exist")
	}
}

func TestAction_Respond(t *testing.T) {
	a := Action{Type: "pet", Response: "You pet the cat!"}

	response := a.Respond("player1", 50)

	if !strings.Contains(response, "player1") {
		t.Errorf("response should contain nick, got %q", response)
	}
	if !strings.Contains(response, "50%") {
		t.Errorf("response should contain love percentage, got %q", response)
	}
}

func TestExecuteAction_Pet_LowLove(t *testing.T) {
	ca := NewCatActions()

	response := ca.ExecuteAction("pet", "player1", "purrito")

	if !strings.Contains(response, "cautiously") {
		t.Errorf("low love pet should be cautious, got %q", response)
	}
	if ca.LoveMeter != 10 {
		t.Errorf("expected love 10 after pet, got %d", ca.LoveMeter)
	}
}

func TestExecuteAction_Pet_HighLove(t *testing.T) {
	ca := NewCatActions()
	ca.LoveMeter = 70 // High love

	response := ca.ExecuteAction("pet", "player1", "purrito")

	if strings.Contains(response, "cautiously") {
		t.Errorf("high love pet should not be cautious, got %q", response)
	}
	if ca.LoveMeter != 80 {
		t.Errorf("expected love 80 after pet, got %d", ca.LoveMeter)
	}
}

func TestExecuteAction_Pet_MaxLove(t *testing.T) {
	ca := NewCatActions()
	ca.LoveMeter = 95 // Near max

	ca.ExecuteAction("pet", "player1", "purrito")

	if ca.LoveMeter != 100 {
		t.Errorf("love should cap at 100, got %d", ca.LoveMeter)
	}
}

func TestExecuteAction_Kick(t *testing.T) {
	ca := NewCatActions()
	ca.LoveMeter = 50
	// Add kick action to map
	ca.Actions["kick"] = Action{Type: "kick", Response: "Kick!"}

	response := ca.ExecuteAction("kick", "player1", "purrito")

	if !strings.Contains(response, "hisses") {
		t.Errorf("kick should cause hiss, got %q", response)
	}
	if ca.LoveMeter != 35 {
		t.Errorf("expected love 35 after kick, got %d", ca.LoveMeter)
	}
}

func TestExecuteAction_Kick_MinLove(t *testing.T) {
	ca := NewCatActions()
	ca.LoveMeter = 10
	// Add kick action to map
	ca.Actions["kick"] = Action{Type: "kick", Response: "Kick!"}

	ca.ExecuteAction("kick", "player1", "purrito")

	if ca.LoveMeter != 0 {
		t.Errorf("love should not go below 0, got %d", ca.LoveMeter)
	}
}

func TestExecuteAction_UnknownAction(t *testing.T) {
	ca := NewCatActions()

	response := ca.ExecuteAction("unknown", "player1", "purrito")

	if response != "Unknown action." {
		t.Errorf("expected 'Unknown action.', got %q", response)
	}
}

func TestExecuteAction_NotPurrito(t *testing.T) {
	ca := NewCatActions()

	response := ca.ExecuteAction("pet", "player1", "someone_else")

	if !strings.Contains(response, "only interact with purrito") {
		t.Errorf("should only allow purrito target, got %q", response)
	}
}

func TestExecuteAction_DefaultCase(t *testing.T) {
	ca := NewCatActions()
	// Add an action that doesn't have special handling
	ca.Actions["wave"] = Action{Type: "wave", Response: "You wave!"}

	response := ca.ExecuteAction("wave", "player1", "purrito")

	if !strings.Contains(response, "doesn't understand") {
		t.Errorf("unhandled action should return confusion, got %q", response)
	}
}

func TestExecuteAction_CaseInsensitive(t *testing.T) {
	ca := NewCatActions()

	response := ca.ExecuteAction("pet", "player1", "PURRITO")

	if strings.Contains(response, "only interact with purrito") {
		t.Error("target should be case insensitive")
	}
}
