package cat_actions

import (
	"testing"
	"time"
)

func TestRemainingCatnip(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		lastUsed time.Time
		want     time.Duration
	}{
		{
			name:     "never used (zero time)",
			now:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			lastUsed: time.Time{},
			want:     0,
		},
		{
			name:     "used 1 hour ago",
			now:      time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC),
			lastUsed: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			want:     23 * time.Hour,
		},
		{
			name:     "used 23 hours ago",
			now:      time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
			lastUsed: time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
			want:     1 * time.Hour,
		},
		{
			name:     "used exactly 24 hours ago",
			now:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			lastUsed: time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
			want:     0,
		},
		{
			name:     "used more than 24 hours ago",
			now:      time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
			lastUsed: time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
			want:     0,
		},
		{
			name:     "used 30 minutes ago",
			now:      time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC),
			lastUsed: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			want:     23*time.Hour + 30*time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := remainingCatnip(tt.now, tt.lastUsed)
			if got != tt.want {
				t.Errorf("remainingCatnip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatRemaining(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "zero",
			d:    0,
			want: "0 m",
		},
		{
			name: "negative (should clamp to 0)",
			d:    -5 * time.Minute,
			want: "0 m",
		},
		{
			name: "30 minutes",
			d:    30 * time.Minute,
			want: "30 m",
		},
		{
			name: "59 minutes",
			d:    59 * time.Minute,
			want: "59 m",
		},
		{
			name: "1 hour exactly",
			d:    1 * time.Hour,
			want: "1 h 0 m",
		},
		{
			name: "1 hour 30 minutes",
			d:    1*time.Hour + 30*time.Minute,
			want: "1 h 30 m",
		},
		{
			name: "23 hours 59 minutes",
			d:    23*time.Hour + 59*time.Minute,
			want: "23 h 59 m",
		},
		{
			name: "5 hours 15 minutes",
			d:    5*time.Hour + 15*time.Minute,
			want: "5 h 15 m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRemaining(tt.d)
			if got != tt.want {
				t.Errorf("formatRemaining(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestNormalizeNick(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple nick",
			input: "player1",
			want:  "player1",
		},
		{
			name:  "uppercase",
			input: "Player1",
			want:  "player1",
		},
		{
			name:  "with leading spaces",
			input: "  player1",
			want:  "player1",
		},
		{
			name:  "with trailing spaces",
			input: "player1  ",
			want:  "player1",
		},
		{
			name:  "with op prefix @",
			input: "@player1",
			want:  "player1",
		},
		{
			name:  "with voice prefix +",
			input: "+player1",
			want:  "player1",
		},
		{
			name:  "with owner prefix ~",
			input: "~player1",
			want:  "player1",
		},
		{
			name:  "with admin prefix &",
			input: "&player1",
			want:  "player1",
		},
		{
			name:  "with halfop prefix %",
			input: "%player1",
			want:  "player1",
		},
		{
			name:  "with multiple prefixes",
			input: "~&@%+player1",
			want:  "player1",
		},
		{
			name:  "uppercase with prefix and spaces",
			input: "  @Player1  ",
			want:  "player1",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only prefixes",
			input: "@+",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeNick(tt.input)
			if got != tt.want {
				t.Errorf("normalizeNick(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCatnipCooldownConstant(t *testing.T) {
	if catnipCooldown != 24*time.Hour {
		t.Errorf("catnipCooldown = %v, want 24h", catnipCooldown)
	}
}
