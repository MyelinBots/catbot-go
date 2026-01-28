package context_manager

import (
	"context"
	"testing"
)

func TestSetNickContext(t *testing.T) {
	ctx := context.Background()
	ctx = SetNickContext(ctx, "TestPlayer")

	nick := GetNickContext(ctx)
	if nick != "testplayer" {
		t.Errorf("expected lowercased nick 'testplayer', got %q", nick)
	}
}

func TestSetNickContext_Lowercase(t *testing.T) {
	ctx := context.Background()
	ctx = SetNickContext(ctx, "player1")

	nick := GetNickContext(ctx)
	if nick != "player1" {
		t.Errorf("expected nick 'player1', got %q", nick)
	}
}

func TestGetNickContext_Empty(t *testing.T) {
	ctx := context.Background()

	nick := GetNickContext(ctx)
	if nick != "" {
		t.Errorf("expected empty nick from fresh context, got %q", nick)
	}
}

func TestSetNickContext_Overwrite(t *testing.T) {
	ctx := context.Background()
	ctx = SetNickContext(ctx, "player1")
	ctx = SetNickContext(ctx, "player2")

	nick := GetNickContext(ctx)
	if nick != "player2" {
		t.Errorf("expected nick 'player2', got %q", nick)
	}
}

func TestNickStruct(t *testing.T) {
	// Ensure Nick{} can be used as a key
	n1 := Nick{}
	n2 := Nick{}

	if n1 != n2 {
		t.Error("two Nick{} instances should be equal")
	}
}
