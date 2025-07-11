package catbot

import (
	"context"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	"math/rand"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
)

// IRCClient defines the interface for IRC client communication
type IRCClient interface {
	Privmsg(channel, message string)
}

// CatBot handles cat actions and message responses
type CatBot struct {
	CatActions cat_actions.CatActionsImpl
	IrcClient  IRCClient
	Channel    string
	times      []int
}

// NewCatBot initializes the CatBot instance
func NewCatBot(client IRCClient, channel string) *CatBot {
	return &CatBot{
		CatActions: cat_actions.NewCatActions(),
		IrcClient:  client,
		Channel:    channel,
		times:      []int{5, 30, 60, 120, 300, 600, 900},
	}
}

// HandleCatCommand processes commands like !pet purrito from users
func (cb *CatBot) HandleCatCommand(ctx context.Context, args ...string) error {
	nick := context_manager.GetNickContext(ctx)
	// message
	parts := strings.Split(args[0], " ")
	if len(parts) < 2 {
		cb.IrcClient.Privmsg(cb.Channel, "Usage: !pet purrito")

		return nil
	}

	action := strings.TrimPrefix(parts[0], "!")
	target := parts[1]

	response := cb.CatActions.ExecuteAction(action, nick, target)
	cb.IrcClient.Privmsg(cb.Channel, response)

	return nil
}

func (cb *CatBot) Start(ctx context.Context) {
	// randomly select a time from the times slice
	for {
		select {
		case <-ctx.Done():
			return
		default:
			cb.HandleRandomAction(ctx)
			randomTime := cb.times[rand.Intn(len(cb.times))]
			<-time.After(time.Duration(randomTime) * time.Second)
		}
	}
}

func (cb *CatBot) HandleRandomAction(ctx context.Context) {
	action := cb.CatActions.GetRandomAction()
	cb.IrcClient.Privmsg(cb.Channel, action)
}
