package context_manager

import (
	"context"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

type Key string

const (
	NickKey Key = "nick"
	LineKey Key = "line"
)

func SetNickContext(ctx context.Context, nick string) context.Context {
	return context.WithValue(ctx, NickKey, strings.ToLower(nick))
}

func GetNickFromContext(ctx context.Context) string {
	nick, ok := ctx.Value(NickKey).(string)
	if !ok {
		return "unknown"
	}
	return nick
}

func SetLineContext(ctx context.Context, line *irc.Line) context.Context {
	return context.WithValue(ctx, LineKey, line)
}

func GetLineFromContext(ctx context.Context) *irc.Line {
	line, ok := ctx.Value(LineKey).(*irc.Line)
	if !ok {
		return nil
	}
	return line
}
