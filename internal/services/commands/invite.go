package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
)

// InviteHandler allows users to invite purrito to their own channels
func InviteHandler(ircClient *irc.Conn) func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		nick := context_manager.GetNickContext(ctx)

		if len(args) < 1 || strings.ToLower(args[0]) != "purrito1" {
			return fmt.Errorf("usage: !invite purrito1")
		}

		// get line from args
		channel := args[1]

		ircClient.Join(channel)
		ircClient.Privmsg(channel, fmt.Sprintf("purrito: meows and joins %s's channel. 🐾", nick))

		fmt.Println("Invite command received from", nick)
		return nil
	}
}
