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

func init() {
	rand.Seed(time.Now().UnixNano())
}

// --------------------------------------------------
// Interfaces
// --------------------------------------------------

type IRCClient interface {
	Privmsg(channel, message string)
}

// รองรับ daily decay แบบมี warning
type dailyDecayerWithWarning interface {
	DailyDecayWithWarning(ctx context.Context) ([]string, error)
}

type dailyDecayer interface {
	DailyDecayAll(ctx context.Context) error
}

// --------------------------------------------------
// CatBot
// --------------------------------------------------

type CatBot struct {
	IrcClient     IRCClient
	CatActions    cat_actions.CatActionsImpl
	Channel       string
	Network       string
	times         []int
	CatPlayerRepo cat_player.CatPlayerRepository

	mu           sync.RWMutex
	presentUntil time.Time
	lastAppear   time.Time
	nextAppear   time.Time
	appearedAt   time.Time
	interacted   bool
}

// --------------------------------------------------
// Constructor
// --------------------------------------------------

func NewCatBot(
	client IRCClient,
	catPlayerRepo cat_player.CatPlayerRepository,
	network, channel string,
) *CatBot {
	return &CatBot{
		IrcClient:     client,
		CatActions:    cat_actions.NewCatActions(catPlayerRepo, network, channel),
		Channel:       channel,
		Network:       network,
		times:         []int{1800}, // 30 นาที 1800 วินาที 5 mins 600 วินาที
		CatPlayerRepo: catPlayerRepo,
	}
}

// --------------------------------------------------
// Presence helpers
// --------------------------------------------------

func (cb *CatBot) consumePresence() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if time.Now().Before(cb.presentUntil) {
		cb.presentUntil = time.Now()
		cb.interacted = true
		return true
	}
	return false
}

func (cb *CatBot) IsPresent() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return time.Now().Before(cb.presentUntil)
}

func (cb *CatBot) AppearTimes() (last, next time.Time) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastAppear, cb.nextAppear
}

// --------------------------------------------------
// Command handler
// --------------------------------------------------

func (cb *CatBot) HandleCatCommand(ctx context.Context, args ...string) error {
	nick := context_manager.GetNickContext(ctx)

	if len(args) == 0 {
		cb.IrcClient.Privmsg(cb.Channel, "Check !purrito for help")
		return nil
	}

	parts := strings.Fields(args[0])
	if len(parts) < 2 {
		cb.IrcClient.Privmsg(cb.Channel, "Check !purrito for help")
		return nil
	}

	action := strings.ToLower(strings.TrimPrefix(parts[0], "!"))
	target := parts[1]

	needsPurritoPresent := map[string]bool{
		"pet":    true,
		"love":   true,
		"feed":   true,
		"catnip": true,
		"laser":  true,
	}

	if needsPurritoPresent[action] && strings.EqualFold(target, "purrito") {
		if !cb.consumePresence() {
			cb.IrcClient.Privmsg(
				cb.Channel,
				"🐾 Purrito is not here right now. Wait until he shows up!",
			)
			return nil
		}
	}

	response := cb.CatActions.ExecuteAction(action, nick, target)
	cb.IrcClient.Privmsg(cb.Channel, response)
	return nil
}

// --------------------------------------------------
// Game loop (ONLY ONE Start)
// --------------------------------------------------

func (cb *CatBot) Start(ctx context.Context) {
	appearTimer := time.NewTimer(0)
	defer appearTimer.Stop()

	decayTicker := time.NewTicker(24 * time.Hour)
	defer decayTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-appearTimer.C:
			cb.HandleRandomAction()

			wait := cb.times[rand.Intn(len(cb.times))]
			cb.mu.Lock()
			cb.nextAppear = time.Now().Add(time.Duration(wait) * time.Second)
			cb.mu.Unlock()

			appearTimer.Reset(time.Duration(wait) * time.Second)

		case <-decayTicker.C:
			if ca, ok := cb.CatActions.(*cat_actions.CatActions); ok {
				if d, ok := any(ca.LoveMeter).(dailyDecayerWithWarning); ok {
					msgs, err := d.DailyDecayWithWarning(context.Background())
					if err == nil {
						for _, m := range msgs {
							cb.IrcClient.Privmsg(cb.Channel, m)
						}
					}
					continue
				}
				if d, ok := any(ca.LoveMeter).(dailyDecayer); ok {
					_ = d.DailyDecayAll(context.Background())
				}
			}
		}
	}
}

// --------------------------------------------------
// Appearance logic
// --------------------------------------------------

func (cb *CatBot) HandleRandomAction() {
	action := cb.CatActions.GetRandomAction()
	cb.IrcClient.Privmsg(cb.Channel, "🐈 meowww ... "+action)

	now := time.Now()

	cb.mu.Lock()
	cb.lastAppear = now
	cb.presentUntil = now.Add(10 * time.Minute)
	cb.appearedAt = now
	cb.interacted = false
	cb.mu.Unlock()

	go func(appearTime time.Time) {
		time.Sleep(10 * time.Minute)

		cb.mu.RLock()
		stillSame := cb.appearedAt.Equal(appearTime)
		quiet := !cb.interacted
		cb.mu.RUnlock()

		if stillSame && quiet {
			cb.IrcClient.Privmsg(
				cb.Channel,
				"(=^‥^=)っ stretches, yawns, and wanders off into the shadows 🐾",
			)
		}
	}(now)
}

// ConsumePresence makes Purrito vanish immediately (same logic as pet/feed/love).
func (cb *CatBot) ConsumePresence() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if time.Now().Before(cb.presentUntil) {
		cb.presentUntil = time.Now()
		cb.interacted = true
		return true
	}
	return false
}
