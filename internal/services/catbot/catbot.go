package catbot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/bondpoints"
	"github.com/MyelinBots/catbot-go/internal/services/bondrewards"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

// --------------------------------------------------
// Interfaces
// --------------------------------------------------

type IRCClient interface {
	Privmsg(channel, message string)
}

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
	CatPlayerRepo cat_player.CatPlayerRepository

	// presence state
	mu           sync.RWMutex
	presentUntil time.Time
	lastAppear   time.Time
	nextAppear   time.Time
	appearedAt   time.Time
	interacted   bool

	// bonded endgame
	BondPoints bondpoints.Service
}

// --------------------------------------------------
// Constructor
// --------------------------------------------------

func NewCatBot(
	client IRCClient,
	catPlayerRepo cat_player.CatPlayerRepository,
	network, channel string,
	spawnWindow, minRespawn, maxRespawn time.Duration,
) *CatBot {
	cb := &CatBot{
		IrcClient:     client,
		CatActions:    cat_actions.NewCatActions(catPlayerRepo, network, channel, spawnWindow, minRespawn, maxRespawn),
		Channel:       channel,
		Network:       network,
		CatPlayerRepo: catPlayerRepo,
		BondPoints:    bondpoints.New(catPlayerRepo),
	}
	return cb
}

// --------------------------------------------------
// Presence helpers
// --------------------------------------------------

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

	// CatActions.ExecuteAction handles presence gating internally
	// (catnip is allowed without presence; other actions require it)
	response := cb.CatActions.ExecuteAction(action, nick, target)

	// append bonded progress for actions targeting purrito
	needsBondProgress := map[string]bool{
		"pet":    true,
		"love":   true,
		"feed":   true,
		"catnip": true,
		"laser":  true,
	}
	if needsBondProgress[action] && strings.EqualFold(target, "purrito") {
		response = cb.appendBondProgress(ctx, nick, response)
	}

	cb.IrcClient.Privmsg(cb.Channel, response)
	return nil
}

// --------------------------------------------------
// Game loop
// --------------------------------------------------

func (cb *CatBot) Start(ctx context.Context) {
	decayTicker := time.NewTicker(24 * time.Hour)
	defer decayTicker.Stop()

	// Presence ticker checks for spawn/leave messages every 10 seconds
	presenceTicker := time.NewTicker(10 * time.Second)
	defer presenceTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-presenceTicker.C:
			// Check for spawn/leave messages
			if ca, ok := cb.CatActions.(*cat_actions.CatActions); ok {
				spawnMsg, leaveMsg := ca.TickPresence()
				if leaveMsg != "" {
					cb.IrcClient.Privmsg(cb.Channel, leaveMsg)
				}
				if spawnMsg != "" {
					cb.IrcClient.Privmsg(cb.Channel, spawnMsg)
				}
			}

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
					if err := d.DailyDecayAll(context.Background()); err != nil {
						log.Printf("daily decay error: %v", err)
					}
				}
			}
		}
	}
}

// --------------------------------------------------
// Bonded helper (formatting + DB reads)
// --------------------------------------------------

// normalizeNick strips IRC prefixes and lowercases the nick
func normalizeNick(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	n = strings.TrimLeft(n, "~&@%+")
	return n
}

func (cb *CatBot) appendBondProgress(ctx context.Context, nick string, msg string) string {
	ca, ok := cb.CatActions.(*cat_actions.CatActions)
	if !ok || ca.LoveMeter == nil {
		return msg
	}

	if ca.LoveMeter.Get(nick) != 100 {
		return msg
	}

	normalizedNick := normalizeNick(nick)

	oldP, _ := cb.CatPlayerRepo.GetPlayerByName(
		ctx,
		normalizedNick,
		ca.Network,
		ca.Channel,
	)
	oldHighest := 0
	if oldP != nil {
		oldHighest = oldP.HighestBondStreak
	}

	res, err := cb.BondPoints.RecordBondedInteraction(
		ctx,
		normalizedNick,
		ca.Network,
		ca.Channel,
	)
	if err != nil {
		return msg
	}

	unlocks := bondrewards.GiftUnlocks(oldHighest, res.HighestStreak)
	if len(unlocks) > 0 {
		mask := 0
		for _, u := range unlocks {
			mask |= u.GiftMask
		}
		_ = cb.CatPlayerRepo.AddGiftsUnlocked(
			ctx,
			normalizedNick,
			ca.Network,
			ca.Channel,
			mask,
		)

		msg += fmt.Sprintf(" :: 😸🎁 %s unlocked", unlocks[0].GiftName)
	}

	title := bondrewards.TitleForHighestStreak(res.HighestStreak)

	if res.AwardedPoints > 0 {
		return msg + fmt.Sprintf(
			" :: Streak: %d day(s) :: +%d BondPoints :: Total: %d :: Title: %s",
			res.Streak, res.AwardedPoints, res.TotalPoints, title,
		)
	}

	return msg + fmt.Sprintf(
		" :: Streak: %d day(s) :: already bonded today :: Total: %d :: Title: %s",
		res.Streak, res.TotalPoints, title,
	)
}
