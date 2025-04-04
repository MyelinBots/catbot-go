package catbot

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
)

// Interfaces
type CatBot interface {
	HandleCatCommand(ctx context.Context, args ...string) error
	Start(ctx context.Context)
	HandleRandomAction(ctx context.Context)
}

type IRCClient interface {
	Privmsg(channel, message string)
}

// Implementation
type CatBotImpl struct {
	LoveMeter  lovemeter.LoveMeter
	CatActions *cat_actions.CatActions
	times      []int
	IrcClient  IRCClient
	Channel    string
	Network    string
}

// Constructor
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

// Handle user command (e.g., !pet purrito)
func (cb *CatBotImpl) HandleCatCommand(ctx context.Context, args ...string) error {
	if len(args) < 2 {
		cb.IrcClient.Privmsg(cb.Channel, "Usage: !pet purrito")
		return nil
	}

	nick := args[0]
	target := args[1]
	extras := args[2:]

	allArgs := append([]string{nick, target}, extras...)
	response := cb.CatActions.ExecuteAction("pet", allArgs...)
	fullMessage := response + " The cat " + cb.GetRandomReaction() + "! (Love meter: " + cb.loveLevelString() + ")"
	cb.IrcClient.Privmsg(cb.Channel, fullMessage)

	return nil
}

// Background random reaction loop
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

// Passive random action
func (cb *CatBotImpl) HandleRandomAction(ctx context.Context) {
	reaction := cb.GetRandomReaction()
	cb.IrcClient.Privmsg(cb.Channel, reaction)
}

// Pick a random cat reaction
func (cb *CatBotImpl) GetRandomReaction() string {
	reactions := []string{
		"meows softly", "purrs", "rubs against your leg", "rolls over", "stares at you lovingly",
	}
	return reactions[rand.Intn(len(reactions))]
}

// Convert love meter to string
func (cb *CatBotImpl) loveLevelString() string {
	return strconv.Itoa(cb.LoveMeter.Get())
}
