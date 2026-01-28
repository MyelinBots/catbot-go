package cat_actions

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/bondpoints"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var emotes = []string{
	"meows happily (=^･^=)",
	"rubs against leg (=^-ω-^=)",
	"purrs warmly (^^=^^)",
	"nuzzles gently (=^･o･^=)ﾉ",
	"flicks its tail playfully (=^･ｪ･^=)",
	"stretches and yawns (=^･ω･^=)ﾉﾞ",
	"rolls over for belly rubs (≧◡≦)ﾉ",
	"gives a soft chirp (=^･ｪ･^=)っ",
	"licks its paw and looks (=^‥^=)",
	"blinks slowly (=^-ᆺ-^=)",
	"purrs contentedly (^・ω・^)ﾉﾞ",
	"curls up beside you (｡♥‿♥｡)",
	"gives a gentle headbutt (=^･ω･^)つ",
	"flicks its ears (^•ﻌ•^)",
	"swishes its tail (≧ω≦)",
	"paws at the air (=^･ｪ･^=)っ",
	"gives a playful swipe (•ω•)",
	"chases a sunbeam (^ↀᴥↀ^)",
	"sniffs curiously (=^･ｪ･^=)",
	"gives a happy meow (=^▽^=)",
	"pounces playfully (=^･ω･^=)つ",
	"gives a soft trill (=^-ω-^=)",
}

var rejects = []string{
	"hisses and moves away (╬ Ò﹏Ó)",
	"growls softly, not in the mood (≖︿≖ )",
	"glares coldly (≧д≦ヾ)",
	"turns their back (￣︿￣)",
	"gives a disdainful look (¬_¬ )",
	"flicks its tail in annoyance (ಠ_ಠ)",
	"lets out a displeased meow (╯^╰)",
	"stiffens and walks away ( =①ω①=)",
	"gives a sharp meow and walks off (＞﹏＜)",
	"scratches the ground and ignores you (=`ω´= )",
	"gives a dismissive flick of the tail (￣へ￣ )",
	"ears flatten in irritation (`･ω･´)っ",
	"gives a warning hiss (ﾒΦ皿Φ)",
	"swats the air and moves away (╬ΦᆺΦ)",
	"gives a disdainful glance (Φ 皿 Φ)",
	"turns its head away (￣ω￣;)",
	"ignores you completely (－‸ლ)",
	"gives a cold stare (ΦωΦ)",
	"flicks its tail and walks away (￣^￣)",
	"lets out an annoyed meow (｀皿´)ノ",
}

type CatActionsImpl interface {
	GetActions() []string
	GetRandomAction() string
	ExecuteAction(actionName, player, target string) string
	IsHere() bool
}

type CatActions struct {
	LoveMeter     lovemeter.LoveMeter
	Actions       []string
	CatPlayerRepo cat_player.CatPlayerRepository
	Network       string
	Channel       string

	mu           sync.RWMutex
	slapWarned   map[string]bool
	catnipUsedAt map[string]time.Time
	BondPoints   bondpoints.Service

	hereUntil time.Time
}

func NewCatActions(catPlayerRepo cat_player.CatPlayerRepository, network, channel string) CatActionsImpl {
	return &CatActions{
		LoveMeter:     lovemeter.NewLoveMeter(catPlayerRepo, network, channel),
		BondPoints:    bondpoints.New(catPlayerRepo),
		Actions:       emotes,
		CatPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,

		slapWarned:   make(map[string]bool),
		catnipUsedAt: make(map[string]time.Time),
	}
}

func (ca *CatActions) GetActions() []string { return ca.Actions }

func (ca *CatActions) GetRandomAction() string {
	return ca.Actions[rand.Intn(len(ca.Actions))]
}

func (ca *CatActions) appendBondProgress(_ string, msg string) string { return msg }

func (ca *CatActions) IsHere() bool {
	ca.mu.RLock()
	defer ca.mu.RUnlock()
	return time.Now().Before(ca.hereUntil)
}

func (ca *CatActions) EnsureHere(forHowLong time.Duration) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	now := time.Now()
	until := now.Add(forHowLong)

	// extend presence if already here
	if until.After(ca.hereUntil) {
		ca.hereUntil = until
	}
}

// --------------------
// Helpers
// --------------------

func normalizeAction(action string) string {
	a := strings.ToLower(strings.TrimSpace(action))
	return strings.TrimPrefix(a, "!")
}

// --------------------
// Catnip cooldown (daily)
// --------------------

func (ca *CatActions) CatnipRemaining(player string) time.Duration {
	key := normalizeNick(player)
	now := time.Now()

	ca.mu.RLock()
	last := ca.catnipUsedAt[key]
	ca.mu.RUnlock()

	return remainingCatnip(now, last)
}

func (ca *CatActions) CatnipOnCooldown(player string) bool { return ca.CatnipRemaining(player) > 0 }

// --------------------
// Presence gate
// --------------------
//
// All actions (pet, love, feed, laser, catnip) require Purrito to be present.
func (ca *CatActions) gatePresenceForAction(action string) (bool, string) {
	// all actions require presence
	if !ca.IsHere() {
		return false, "🐾 Purrito is not here right now. Wait until he shows up!"
	}
	return true, ""
}

func (ca *CatActions) ExecuteAction(actionName, player, target string) string {
	t := normalizeAction(target)
	a := normalizeAction(actionName)

	if t != "purrito" {
		pt := cases.Title(language.English).String(target)
		otherRejects := []string{
			fmt.Sprintf("%s, %s does not want to be pet.", player, pt),
			fmt.Sprintf("%s, purrito looks confused… why are you petting %s?", player, pt),
			fmt.Sprintf("%s, purrito meows awkwardly. Only purrito likes being pet not %s...", player, pt),
			fmt.Sprintf("%s, %s is not interested in being pet… purrito is waiting!", player, pt),
		}
		return otherRejects[rand.Intn(len(otherRejects))]
	}

	switch a {
	case "pet", "love":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}
		if rand.Intn(100) < 95 {
			ca.LoveMeter.Increase(player, 1)
			return ca.acceptMessage(player)
		}
		ca.LoveMeter.Decrease(player, 1)
		return ca.rejectMessage(player)

	case "feed":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}
		foods := []string{"salmon", "tuna", "sardines", "chicken", "kibble", "milk", "fish snacks", "cream", "shrimp", "turkey", "beef"}
		food := foods[rand.Intn(len(foods))]

		if rand.Intn(100) < 60 {
			ca.LoveMeter.Increase(player, 1)
			return ca.feedAcceptMessage(player, food)
		}
		ca.LoveMeter.Decrease(player, 1)
		return ca.feedRejectMessage(player, food)

	case "laser":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}
		// 60% accept, 40% reject - love change happens here
		if rand.Intn(100) < 60 {
			ca.LoveMeter.Increase(player, 1)
			return ca.laserAcceptMessage(player)
		}
		ca.LoveMeter.Decrease(player, 1)
		return ca.laserRejectMessage(player)

	case "status":
		return ca.statusMessage(player)

	case "catnip":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}

		return ca.catnipMessage(player)

	case "slap", "kick":
		// (optional) if you want slap/kick to require presence too:
		// if ok, msg := ca.requireHere(a); !ok { return msg }

		key := strings.ToLower(strings.TrimSpace(player))

		ca.mu.Lock()
		warned := ca.slapWarned[key]
		if !warned {
			ca.slapWarned[key] = true
		}
		ca.mu.Unlock()

		if !warned {
			firstWarnings := []string{
				fmt.Sprintf("😾 Purrito flattens his ears at %s... This is your warning... do not slap him again...", player),
				fmt.Sprintf("⚠️ Purrito stares at %s with shocked eyes… he did not like that...", player),
				fmt.Sprintf("😿 Purrito backs away from %s… please be gentle with him.", player),
				fmt.Sprintf("⚠️ Purrito watches %s carefully… one more slap and he will be upset.", player),
				fmt.Sprintf("😼 Purrito lifts a paw at %s in warning… do not try that again...", player),
			}
			return firstWarnings[rand.Intn(len(firstWarnings))]
		}

		ca.LoveMeter.Decrease(player, 1)
		love := ca.LoveMeter.Get(player)
		mood := ca.LoveMeter.GetMood(player)
		bar := ca.LoveMeter.GetLoveBar(player)

		secondPunishments := []string{
			fmt.Sprintf("😾 Purrito swats back at %s and looks hurt. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😾 Purrito hisses softly at %s… his heart hurts. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😿 Purrito lowers his ears… %s made him sad. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😿 Purrito looks betrayed by %s. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😾 Purrito steps back from %s… do not hurt him. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
		}
		return secondPunishments[rand.Intn(len(secondPunishments))]

	default:
		return "purrito tilts its head, not sure what you mean 🐾"
	}
}

// --------------------
// Messages (unchanged)
// --------------------

func (ca *CatActions) acceptMessage(player string) string {
	emote := emotes[rand.Intn(len(emotes))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	base := fmt.Sprintf("%s at %s and your love meter is now %d%% and purrito is now %s %s",
		emote, player, love, mood, bar)

	return ca.appendBondProgress(player, base)
}

func (ca *CatActions) rejectMessage(player string) string {
	reject := rejects[rand.Intn(len(rejects))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	base := fmt.Sprintf("purrito %s at %s and your love meter is now %d%% and purrito is now %s %s",
		reject, player, love, mood, bar)

	return ca.appendBondProgress(player, base)
}

func (ca *CatActions) feedAcceptMessage(player, food string) string {
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	lines := []string{
		fmt.Sprintf("😺 Purrito happily munches the %s you gave, %s! Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
		fmt.Sprintf("😻 Purrito devours the %s and purrs loudly at %s. Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
		fmt.Sprintf("🍣 Purrito LOVES the %s from %s. Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
	}
	return ca.appendBondProgress(player, lines[rand.Intn(len(lines))])
}

func (ca *CatActions) feedRejectMessage(player, food string) string {
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	lines := []string{
		fmt.Sprintf("😼 Purrito sniffs the %s from %s and turns away... your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
		fmt.Sprintf("😾 Purrito refuses the %s. %s, he is a picky cat. Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
		fmt.Sprintf("🙀 Purrito looks offended by the %s from %s. Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
	}
	return ca.appendBondProgress(player, lines[rand.Intn(len(lines))])
}

func (ca *CatActions) laserAcceptMessage(player string) string {
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	lines := []string{
		fmt.Sprintf("🔦⚡️ The laser flickers! Purrito darts after it, paws flying everywhere! Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦⚡️ Purrito spots the laser and wiggles... then pounces! Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦⚡️ Purrito chases the laser dot in circles... dizzy but happy! Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦⚡️ Purrito dives at the laser, misses, then looks proud anyway. Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦⚡️ The red dot dances... Purrito bats at it with lightning speed! Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
	}
	return ca.appendBondProgress(player, lines[rand.Intn(len(lines))])
}

func (ca *CatActions) laserRejectMessage(player string) string {
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	lines := []string{
		fmt.Sprintf("🔦😾 Purrito narrows his eyes... not impressed by the laser right now. Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦🙄 Purrito ignores the dot and grooms his paw instead. Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦😿 Purrito flops down... too tired to chase today. Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦😼 Purrito watches... then turns away like it's beneath him. Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
		fmt.Sprintf("🔦😾 Purrito swishes his tail in annoyance and refuses to play. Your love meter is now %d%% and purrito is now %s %s", love, mood, bar),
	}
	return ca.appendBondProgress(player, lines[rand.Intn(len(lines))])
}

func (ca *CatActions) statusMessage(player string) string {
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	return fmt.Sprintf("Purrito status for %s and your love meter is %d%% and purrito is now %s %s",
		player, love, mood, bar)
}

func (ca *CatActions) catnipMessage(player string) string {
	key := normalizeNick(player)
	now := time.Now()

	ca.mu.Lock()
	last := ca.catnipUsedAt[key]
	if rem := remainingCatnip(now, last); rem > 0 {
		ca.mu.Unlock()
		return fmt.Sprintf("aww %s, you already used catnip today. Try again in %s.", player, formatRemaining(rem))
	}
	ca.catnipUsedAt[key] = now
	ca.mu.Unlock()

	ca.EnsureHere(30 * time.Minute)

	if rand.Intn(100) < 70 {
		ca.LoveMeter.Increase(player, 3)
		love := ca.LoveMeter.Get(player)
		mood := ca.LoveMeter.GetMood(player)
		bar := ca.LoveMeter.GetLoveBar(player)

		variants := []string{
			fmt.Sprintf("🌿😺 Purrito sniffs the catnip and flops over, rolling around happily at %s... your love meter is now %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("🌿😻 Purrito licks the catnip and goes into hyper-purr mode around %s... your love meter is now %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("🌿🐾 Purrito cuddles into the catnip near %s and purrs loudly... your love meter is now %d%% and purrito is now %s %s", player, love, mood, bar),
		}
		return ca.appendBondProgress(player, variants[rand.Intn(len(variants))])
	}

	ca.LoveMeter.Decrease(player, 1)
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	variants := []string{
		fmt.Sprintf("🌿🙀 Purrito gets overwhelmed by the catnip from %s and needs space. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
		fmt.Sprintf("🌿😾 Purrito sneezes and backs away from %s's catnip... too strong! your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
		fmt.Sprintf("🌿😿 Purrito looks displeased with the catnip from %s and walks off... your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
	}
	return ca.appendBondProgress(player, variants[rand.Intn(len(variants))])
}
