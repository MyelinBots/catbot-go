package catbot

import (
	"context"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
	"math/rand"
	"time"
)

type CatBot interface {
	HandleCatCommand(ctx context.Context, args ...string) error
	Start(ctx context.Context)
	HandleRandomAction(ctx context.Context)
}

type IRCClient interface {
	Privmsg(channel, message string)
}

// CatBotImpl represents the bot functionality
type CatBotImpl struct {
	LoveMeter  lovemeter.LoveMeter
	CatActions cat_actions.CatActions
	times      []int
	IrcClient  IRCClient
	Channel    string
	Network    string
}

func NewCatBot(client IRCClient, network string, channel string) CatBot {
	return &CatBotImpl{
		LoveMeter:  lovemeter.NewLoveMeter(),
		CatActions: cat_actions.NewCatActions(),
		Channel:    channel,
		Network:    network,
		IrcClient:  client,
		// time in seconds to wait, up to 15 minutes
		times: []int{5, 30, 60, 120, 300, 600, 900},
	}
}

func (cb *CatBotImpl) HandleCatCommand(ctx context.Context, args ...string) error {
	cb.IrcClient.Privmsg(cb.Channel, "Meow!")
	return nil
}

func (cb *CatBotImpl) Start(ctx context.Context) {
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

func (cb *CatBotImpl) HandleRandomAction(ctx context.Context) {
	action := cb.CatActions.GetRandomAction()
	cb.IrcClient.Privmsg(cb.Channel, action)
}
