package catbot

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

func init() { rand.Seed(time.Now().UnixNano()) }

// IRC client shim
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
	presentUntil time.Time // Purrito is "present" until this time (gates !pet/!love)
	lastAppear   time.Time // last time Purrito appeared
	nextAppear   time.Time // scheduled next appearance
	appearedAt   time.Time // start time of the current appearance window
	interacted   bool      // flipped true if !pet or !laser happened since appearedAt
}

// NewCatBot initializes the CatBot instance
func NewCatBot(client IRCClient, catPlayerRepo cat_player.CatPlayerRepository, network, channel string) *CatBot {
	return &CatBot{
		IrcClient:     client,
		CatActions:    cat_actions.NewCatActions(catPlayerRepo, network, channel),
		Channel:       channel,
		Network:       network,
		times:         []int{1800}, // ~ every 30 minutes
		CatPlayerRepo: catPlayerRepo,
	}
}

// consumePresence allows exactly one interaction during the window
func (cb *CatBot) consumePresence() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if time.Now().Before(cb.presentUntil) {
		// mark as interacted
		cb.presentUntil = time.Now()
		return true
	}
	return false
}

// IsPresent reports whether Purrito is currently visible (within presence window).
func (cb *CatBot) IsPresent() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return time.Now().Before(cb.presentUntil)
}

// AppearTimes returns the last and next appearance times (thread-safe).
func (cb *CatBot) AppearTimes() (last, next time.Time) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastAppear, cb.nextAppear
}

func (cb *CatBot) MarkInteracted() {
	cb.mu.Lock()
	cb.interacted = true
	cb.mu.Unlock()
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

	rawAction := strings.TrimPrefix(parts[0], "!")
	action := strings.ToLower(rawAction)
	target := parts[1]

	// commands that require Purrito to be present
	needsPurritoPresent := map[string]bool{
		"pet":    true,
		"love":   true,
		"feed":   true,
		"catnip": true,
	}

	// If the command requires Purrito to be present and the target is purrito
	if needsPurritoPresent[action] && strings.EqualFold(target, "purrito") {
		if !cb.consumePresence() {
			// Someone else already interacted or the 10-minute window has passed
			cb.IrcClient.Privmsg(cb.Channel, "🐾 Purrito is not here right now. Wait until he shows up!")
			return nil
		}
		// This user is the one who "made it" in this round → count as successful interaction
		cb.MarkInteracted()
	}

	// Forward to CatActions to generate the response message
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
			cb.HandleRandomAction() // Purrito appears now
			// Sleep until next appearance (~30 min)
			wait := cb.times[rand.Intn(len(cb.times))]
			time.Sleep(time.Duration(wait) * time.Second)
		}
	}
}

// HandleRandomAction: Purrito appears, stays 10 minutes (message if no interaction), presence is 5 minutes for !pet/!love
func (cb *CatBot) HandleRandomAction() {
	action := cb.CatActions.GetRandomAction()
	cb.IrcClient.Privmsg(cb.Channel, "🐈 meowww ... "+action)

	now := time.Now()

	cb.mu.Lock()
	cb.lastAppear = now
	cb.nextAppear = now.Add(30 * time.Minute)   // cadence
	cb.presentUntil = now.Add(10 * time.Minute) // interactable 10m
	cb.appearedAt = now
	cb.interacted = false
	cb.mu.Unlock()

	// After 10 minutes from this appearance, if nobody interacted, post the “wanders off” line.
	go func(appearTime time.Time) {
		<-time.After(10 * time.Minute)
		cb.mu.RLock()
		stillSameAppear := cb.appearedAt.Equal(appearTime)
		quiet := !cb.interacted
		cb.mu.RUnlock()

		if stillSameAppear && quiet {
			cb.IrcClient.Privmsg(cb.Channel, "(=^‥^=)っ ... stretches, yawns, and wanders off into the shadows ... 🐾 (=^‥^=)っ")
		}
	}(now)
}
