package catbot

import (
	"context"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
)

// IRCClient defines the interface for IRC client communication
type IRCClient interface {
	Privmsg(channel, message string)
}

// CatBot handles cat actions and message responses
type CatBot struct {
	CatActions *cat_actions.CatActions
	IrcClient  IRCClient
	Channel    string
}

// NewCatBot initializes the CatBot instance
func NewCatBot(client IRCClient, channel string) *CatBot {
	return &CatBot{
		CatActions: cat_actions.NewCatActions(),
		IrcClient:  client,
		Channel:    channel,
	}
}

// HandleCatCommand processes commands like !pet purrito from users
func (cb *CatBot) HandleCatCommand(ctx context.Context, player string, message string) {
	parts := strings.Fields(message)
	if len(parts) < 2 {
		cb.IrcClient.Privmsg(cb.Channel, "Usage: !pet purrito")
		return
	}

	action := strings.TrimPrefix(parts[0], "!")
	target := parts[1]

	response := cb.CatActions.ExecuteAction(action, player, target)
	cb.IrcClient.Privmsg(cb.Channel, response)
}
