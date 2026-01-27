package cat_actions

import (
	"fmt"
	"strings"
	"time"
)

const catnipCooldown = 24 * time.Hour

func remainingCatnip(now, lastUsed time.Time) time.Duration {
	if lastUsed.IsZero() {
		return 0
	}
	next := lastUsed.Add(catnipCooldown)
	if now.Before(next) {
		return next.Sub(now)
	}
	return 0
}

func formatRemaining(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%d h %d m", h, m)
	}
	return fmt.Sprintf("%d m", m)
}

func normalizeNick(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	n = strings.TrimLeft(n, "~&@%+")
	return n
}
