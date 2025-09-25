package catbot

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

// Seed RNG once for this package
func init() { rand.Seed(time.Now().UnixNano()) }

// IRCClient defines the interface for IRC client communication
type IRCClient interface {
	Privmsg(channel, message string)
}

// CatBot handles cat actions and message responses
type CatBot struct {
	IrcClient     IRCClient
	CatActions    cat_actions.CatActionsImpl
	Channel       string
	Network       string
	times         []int
	CatPlayerRepo cat_player.CatPlayerRepository

	mu           sync.RWMutex
	presentUntil time.Time // Purrito is "present" until this time
}

// NewCatBot initializes the CatBot instance
func NewCatBot(client IRCClient, catPlayerRepo cat_player.CatPlayerRepository, network, channel string) *CatBot {
	return &CatBot{
		IrcClient:     client,
		CatActions:    cat_actions.NewCatActions(catPlayerRepo, network, channel),
		Channel:       channel,
		Network:       network,
		times:         []int{1800}, // 30m
		CatPlayerRepo: catPlayerRepo,
	}
}

func (cb *CatBot) setPresentWindow(d time.Duration) {
	cb.mu.Lock()
	cb.presentUntil = time.Now().Add(d)
	cb.mu.Unlock()
}

func (cb *CatBot) consumePresence() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if time.Now().Before(cb.presentUntil) {
		cb.presentUntil = time.Now() // consume presence
		return true
	}
	return false
}

// HandleCatCommand processes commands like "!pet purrito" from users
func (cb *CatBot) HandleCatCommand(ctx context.Context, args ...string) error {
	nick := context_manager.GetNickContext(ctx)

	if len(args) == 0 {
		cb.IrcClient.Privmsg(cb.Channel, "Usage: !pet purrito")
		return nil
	}
	parts := strings.Fields(args[0])
	if len(parts) < 2 {
		cb.IrcClient.Privmsg(cb.Channel, "Usage: !pet purrito")
		return nil
	}

	action := strings.TrimPrefix(parts[0], "!")
	target := parts[1]

	// Gate petting/love to presence window (one-shot)
	if (action == "pet" || action == "love") && strings.EqualFold(target, "purrito") {
		if !cb.consumePresence() {
			cb.IrcClient.Privmsg(cb.Channel, "🐾 Purrito is not here right now. Wait until he shows up!")
			return nil
		}
	}

	response := cb.CatActions.ExecuteAction(action, nick, target)
	cb.IrcClient.Privmsg(cb.Channel, response)
	return nil
}

func (cb *CatBot) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			cb.HandleRandomAction()
			randomTime := cb.times[rand.Intn(len(cb.times))]
			time.Sleep(time.Duration(randomTime) * time.Second)
		}
	}
}

// HandleRandomAction makes Purrito appear and sets presence window (5 minutes)
func (cb *CatBot) HandleRandomAction() {
	action := cb.CatActions.GetRandomAction()
	cb.IrcClient.Privmsg(cb.Channel, fmt.Sprintf("🐈 meowww ... %s", action))
	cb.setPresentWindow(5 * time.Minute) // interactable for 5 minutes after appearing
}
