// internal/services/commands/toplove.go
package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/user"
	irc "github.com/fluffle/goirc/client"
)

// CommandHandler signature should match your existing controller’s expectation.
// If your CommandController uses a different type, adjust the signature accordingly.
type CommandHandler func(ctx context.Context, line *irc.Line) error

// TopLoveHandler returns a handler that prints a leaderboard to the channel / user.
func TopLoveHandler(userRepo user.UserRepository, conn *irc.Conn) CommandHandler {
	return func(ctx context.Context, line *irc.Line) error {
		raw := line.Args[1] // full message, e.g. "!toplove 10"
		args := strings.Fields(raw)

		limit := 5
		if len(args) > 1 {
			if v, err := strconv.Atoi(args[1]); err == nil {
				if v < 1 {
					v = 1
				}
				if v > 20 {
					v = 20
				}
				limit = v
			}
		}

		users, err := userRepo.TopLoveMeter(limit)
		if err != nil {
			conn.Privmsg(line.Args[0], "❌ error: unable to fetch toplove")
			return err
		}
		if len(users) == 0 {
			conn.Privmsg(line.Args[0], "No love data yet. Try petting the cat first! 😺")
			return nil
		}

		// Build one compact line: #1 Nick (♥ 42), #2 ...
		var b strings.Builder
		fmt.Fprintf(&b, "💖 Top Lovers (Top %d): ", len(users))
		for i, u := range users {
			if i > 0 {
				b.WriteString("  •  ")
			}
			// Adjust u.LoveScore/u.LoveMeter.Score depending on your model
			score := u.LoveScore
			fmt.Fprintf(&b, "#%d %s (♥ %d)", i+1, u.Nickname, score)
		}

		target := line.Args[0] // channel or private nick
		if !strings.HasPrefix(target, "#") {
			target = line.Nick // if PM, reply to nick
		}
		conn.Privmsg(target, b.String())
		return nil
	}
}
