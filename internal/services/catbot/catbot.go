package catbot

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/bondpoints"
	"github.com/MyelinBots/catbot-go/internal/services/bondrewards"
	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

func init() { rand.Seed(time.Now().UnixNano()) }

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
// Presence tuning
// --------------------------------------------------

const (
	appearEverySeconds = 300             // 5 minutes
	presenceDuration   = 3 * time.Minute // stays 3 minutes
)

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
) *CatBot {
	cb := &CatBot{
		IrcClient:     client,
		CatActions:    cat_actions.NewCatActions(catPlayerRepo, network, channel),
		Channel:       channel,
		Network:       network,
		times:         []int{appearEverySeconds},
		CatPlayerRepo: catPlayerRepo,
		BondPoints:    bondpoints.New(catPlayerRepo),
	}
	return cb
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

	// must be present AND consume (vanish immediately)
	if needsPurritoPresent[action] && strings.EqualFold(target, "purrito") {
		if !cb.consumePresence() {
			cb.IrcClient.Privmsg(cb.Channel, "🐾 Purrito is not here right now. Wait until he shows up!")
			return nil
		}
	}

	// execute action -> message contains love/mood/bar already
	response := cb.CatActions.ExecuteAction(action, nick, target)

	// append bonded progress ONLY for: !pet !love !feed !catnip !laser (target purrito)
	if needsPurritoPresent[action] && strings.EqualFold(target, "purrito") {
		response = cb.appendBondProgress(ctx, nick, response)
	}

	cb.IrcClient.Privmsg(cb.Channel, response)
	return nil
}

// --------------------------------------------------
// Game loop
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
	cb.presentUntil = now.Add(presenceDuration) // stays 3 minutes
	cb.appearedAt = now
	cb.interacted = false
	cb.mu.Unlock()

	go func(appearTime time.Time) {
		time.Sleep(presenceDuration)

		cb.mu.RLock()
		stillSame := cb.appearedAt.Equal(appearTime)
		quiet := !cb.interacted
		cb.mu.RUnlock()

		if stillSame && quiet {
			cb.IrcClient.Privmsg(cb.Channel, "(=^‥^=)っ stretches, yawns, and wanders off into the shadows 🐾")
		}
	}(now)
}

// ConsumePresence makes Purrito vanish immediately.
func (cb *CatBot) ConsumePresence() bool { return cb.consumePresence() }

// --------------------------------------------------
// Bonded helper (formatting + DB reads)
// --------------------------------------------------

func (cb *CatBot) appendBondProgress(ctx context.Context, nick string, msg string) string {
	ca, ok := cb.CatActions.(*cat_actions.CatActions)
	if !ok || ca.LoveMeter == nil {
		return msg
	}

	// gate: must be bonded now
	if ca.LoveMeter.Get(nick) != 100 {
		return msg
	}

	// capture old highest before updating (for unlock-diff)
	oldP, _ := cb.CatPlayerRepo.GetPlayerByName(ctx, nick, ca.Network, ca.Channel)
	oldHighest := 0
	if oldP != nil {
		oldHighest = oldP.HighestBondStreak
	}

	// record bondpoints (once per NY day). No “already earned today” message.
	res, err := cb.BondPoints.RecordBondedInteraction(ctx, nick, ca.Network, ca.Channel)
	if err != nil {
		return msg
	}

	// secret gifts unlock (based on HighestStreak progression)
	unlocks := bondrewards.GiftUnlocks(oldHighest, res.HighestStreak)
	if len(unlocks) > 0 {
		mask := 0
		for _, u := range unlocks {
			mask |= u.GiftMask
		}
		// you need repo method: AddGiftsUnlocked(..., mask)
		_ = cb.CatPlayerRepo.AddGiftsUnlocked(ctx, nick, ca.Network, ca.Channel, mask)

		// “เนียน” = show only the first unlocked name
		msg += fmt.Sprintf(" | 🎁 %s unlocked", unlocks[0].GiftName)
	}

	// cute title (mystical/pastel vibe)
	title := bondrewards.TitleForHighestStreak(res.HighestStreak)

	// ✅ format rules you asked:
	// 1) count streak
	// 2) bondpoints independent from lovemeter (only gated by love==100)
	// 3) no “already earned today”
	// 4) still have Total always
	if res.AwardedPoints > 0 {
		return msg + fmt.Sprintf(" | Streak: %d day(s) | +%d BondPoints | Total: %d | Title: %s",
			res.Streak, res.AwardedPoints, res.TotalPoints, title,
		)
	}
	return msg + fmt.Sprintf(" | Streak: %d day(s) | Total: %d | Title: %s",
		res.Streak, res.TotalPoints, title,
	)
}
