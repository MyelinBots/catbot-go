package bondrewards

import (
	"testing"
)

func TestTitleForHighestStreak(t *testing.T) {
	tests := []struct {
		name     string
		highest  int
		expected string
	}{
		{"zero streak", 0, "Soft New Moon 🌙"},
		{"1 day", 1, "Soft New Moon 🌙"},
		{"6 days", 6, "Soft New Moon 🌙"},
		{"7 days exactly", 7, "Moon-Touched Friend 🌙🐾"},
		{"10 days", 10, "Moon-Touched Friend 🌙🐾"},
		{"13 days", 13, "Moon-Touched Friend 🌙🐾"},
		{"14 days exactly", 14, "Starlight Companion ✨🐱"},
		{"20 days", 20, "Starlight Companion ✨🐱"},
		{"29 days", 29, "Starlight Companion ✨🐱"},
		{"30 days exactly", 30, "Lunar Bonded Soul 🌕💫"},
		{"45 days", 45, "Lunar Bonded Soul 🌕💫"},
		{"59 days", 59, "Lunar Bonded Soul 🌕💫"},
		{"60 days exactly", 60, "Keeper of the Night Purr 🌌🐾"},
		{"80 days", 80, "Keeper of the Night Purr 🌌🐾"},
		{"99 days", 99, "Keeper of the Night Purr 🌌🐾"},
		{"100 days exactly", 100, "Eternal Moonbound 🌑♾️"},
		{"200 days", 200, "Eternal Moonbound 🌑♾️"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TitleForHighestStreak(tt.highest)
			if got != tt.expected {
				t.Errorf("TitleForHighestStreak(%d) = %q, want %q", tt.highest, got, tt.expected)
			}
		})
	}
}

func TestGiftUnlocks(t *testing.T) {
	tests := []struct {
		name       string
		oldHighest int
		newHighest int
		wantCount  int
		wantMasks  []int
		wantNames  []string
	}{
		{
			name:       "no unlock - below 7",
			oldHighest: 0,
			newHighest: 5,
			wantCount:  0,
		},
		{
			name:       "unlock 7 day gift",
			oldHighest: 0,
			newHighest: 7,
			wantCount:  1,
			wantMasks:  []int{Gift7},
			wantNames:  []string{"🔔🌙 Pastel Moon Bell"},
		},
		{
			name:       "unlock 14 day gift only",
			oldHighest: 7,
			newHighest: 14,
			wantCount:  1,
			wantMasks:  []int{Gift14},
			wantNames:  []string{"🎀✨ Starlit Ribbon Collar"},
		},
		{
			name:       "unlock multiple gifts at once",
			oldHighest: 0,
			newHighest: 30,
			wantCount:  3,
			wantMasks:  []int{Gift7, Gift14, Gift30},
			wantNames:  []string{"🔔🌙 Pastel Moon Bell", "🎀✨ Starlit Ribbon Collar", "🌕💎 Lunar Memory Charm"},
		},
		{
			name:       "no new unlocks - already had them",
			oldHighest: 30,
			newHighest: 50,
			wantCount:  0,
		},
		{
			name:       "unlock 30 day gift only",
			oldHighest: 14,
			newHighest: 30,
			wantCount:  1,
			wantMasks:  []int{Gift30},
			wantNames:  []string{"🌕💎 Lunar Memory Charm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unlocks := GiftUnlocks(tt.oldHighest, tt.newHighest)
			if len(unlocks) != tt.wantCount {
				t.Errorf("GiftUnlocks(%d, %d) returned %d unlocks, want %d",
					tt.oldHighest, tt.newHighest, len(unlocks), tt.wantCount)
				return
			}

			for i, u := range unlocks {
				if u.GiftMask != tt.wantMasks[i] {
					t.Errorf("unlock[%d].GiftMask = %d, want %d", i, u.GiftMask, tt.wantMasks[i])
				}
				if u.GiftName != tt.wantNames[i] {
					t.Errorf("unlock[%d].GiftName = %q, want %q", i, u.GiftName, tt.wantNames[i])
				}
			}
		})
	}
}

func TestJoinGifts(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		expected string
	}{
		{
			name:     "empty list",
			list:     []string{},
			expected: "None yet — Purrito is watching quietly",
		},
		{
			name:     "nil list",
			list:     nil,
			expected: "None yet — Purrito is watching quietly",
		},
		{
			name:     "single gift",
			list:     []string{"Pastel Moon Bell"},
			expected: "Pastel Moon Bell",
		},
		{
			name:     "multiple gifts",
			list:     []string{"Pastel Moon Bell", "Starlit Ribbon Collar"},
			expected: "Pastel Moon Bell, Starlit Ribbon Collar",
		},
		{
			name:     "three gifts",
			list:     []string{"Gift1", "Gift2", "Gift3"},
			expected: "Gift1, Gift2, Gift3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinGifts(tt.list)
			if got != tt.expected {
				t.Errorf("JoinGifts(%v) = %q, want %q", tt.list, got, tt.expected)
			}
		})
	}
}

func TestGiftConstants(t *testing.T) {
	// Ensure gift constants are distinct bits
	if Gift7&Gift14 != 0 {
		t.Error("Gift7 and Gift14 overlap")
	}
	if Gift7&Gift30 != 0 {
		t.Error("Gift7 and Gift30 overlap")
	}
	if Gift14&Gift30 != 0 {
		t.Error("Gift14 and Gift30 overlap")
	}

	// Ensure they're powers of 2
	if Gift7 != 1 {
		t.Errorf("Gift7 = %d, want 1", Gift7)
	}
	if Gift14 != 2 {
		t.Errorf("Gift14 = %d, want 2", Gift14)
	}
	if Gift30 != 4 {
		t.Errorf("Gift30 = %d, want 4", Gift30)
	}
}
