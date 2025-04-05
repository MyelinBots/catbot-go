package context_manager

import (
	"context"
	"strings"
)

type nickKey struct{}

// SetNickContext stores the nickname into context
func SetNickContext(ctx context.Context, nick string) context.Context {
	return context.WithValue(ctx, nickKey{}, strings.ToLower(nick))
}

// GetNickFromContext retrieves the nickname from context
func GetNickFromContext(ctx context.Context) string {
	nick, ok := ctx.Value(nickKey{}).(string)
	if !ok {
		return "unknown"
	}
	return nick
}
