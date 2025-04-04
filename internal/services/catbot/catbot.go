package catbot

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
)

type CatBot interface {
	HandleCatCommand(ctx context.Context, args ...string) error
	Start(ctx context.Context)
	HandleRandomAction(ctx context.Context)
}

type IRCClient interface {
	Privmsg(channel, message string)
}

type CatBotImpl struct {
	LoveMeter  lovemeter.LoveMeter
	CatActions *cat_actions.CatActions
	times      []int
	IrcClient  IRCClient
	Channel    string
	Network    string
}

func NewCatBot(client IRCClient, network string, channel string) CatBot {
	rand.Seed(time.Now().UnixNano())
	loveMeter := lovemeter.NewLoveMeter()
	return &CatBotImpl{
		LoveMeter:  loveMeter,
		CatActions: cat_actions.NewCatActions(loveMeter),
		Channel:    channel,
		Network:    network,
		IrcClient:  client,
		times:      []int{5, 30, 60, 120, 300, 600, 900},
	}
}

func (cb *CatBotImpl) HandleCatCommand(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		cb.IrcClient.Privmsg(cb.Channel, "Meow? Try !pet purrito.")
		return nil
	}

	command := strings.ToLower(args[0])
	switch command {
	case "!pet":
		response := cb.CatActions.ExecuteAction("pet")
		reaction := cb.GetRandomReaction()
		fullMessage := response + " The cat " + reaction + "! (Love meter: " + cb.loveLevelString() + ")"
		cb.IrcClient.Privmsg(cb.Channel, fullMessage)
	case "!love":
		love := cb.LoveMeter.Get()
		cb.IrcClient.Privmsg(cb.Channel, cb.getLoveMessage(love))
	default:
		cb.IrcClient.Privmsg(cb.Channel, "purrito tilts its head. Unknown command. 🐱")
	}

	return nil
}

func (cb *CatBotImpl) loveLevelString() string {
	return strconv.Itoa(cb.LoveMeter.Get())
}

func (cb *CatBotImpl) getLoveMessage(love int) string {
	switch {
	case love >= 90:
		return "purrito absolutely adores you! 💖"
	case love >= 70:
		return "purrito really likes you! 🐾"
	case love >= 40:
		return "purrito is curious and friendly. 🐱"
	default:
		return "purrito keeps a cautious distance... 😼"
	}
}

func (cb *CatBotImpl) Start(ctx context.Context) {
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
	reaction := cb.GetRandomReaction()
	cb.IrcClient.Privmsg(cb.Channel, reaction)
}

// You can optionally extract this from CatActions instead
func (cb *CatBotImpl) GetRandomReaction() string {
	reactions := []string{
		"meows softly", "purrs", "rubs against your leg", "rolls over", "stares at you lovingly",
	}
	return reactions[rand.Intn(len(reactions))]
}
