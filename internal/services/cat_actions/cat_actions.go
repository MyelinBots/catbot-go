package cat_actions

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type CatActionsImpl interface {
	GetActions() []string
	GetRandomAction() string
	ExecuteAction(actionName, player, target string) string
}

type CatActions struct {
	LoveMeter     lovemeter.LoveMeter
	Actions       []string
	CatPlayerRepo cat_player.CatPlayerRepository
	Network       string
	Channel       string

	mu           sync.RWMutex
	slapWarned   map[string]bool // track who already got a slap warning
	catnipUsedAt map[string]time.Time
}

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
	"gives a sharp meow and walks off (＞﹏＜)",
	"ignores you completely (－‸ლ)",
	"gives a cold stare (ΦωΦ)",
	"flicks its tail and walks away (￣^￣)",
	"lets out an annoyed meow (｀皿´)ノ",
}

// NewCatActions returns a new instance of CatActions
func NewCatActions(catPlayerRepo cat_player.CatPlayerRepository, network, channel string) CatActionsImpl {
	return &CatActions{
		LoveMeter:     lovemeter.NewLoveMeter(catPlayerRepo, network, channel),
		Actions:       emotes,
		CatPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,
		slapWarned:    make(map[string]bool),
		catnipUsedAt:  make(map[string]time.Time),
	}
}

func sameDayNY(a, b time.Time) bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}
	aa := a.In(loc)
	bb := b.In(loc)
	return aa.Year() == bb.Year() && aa.YearDay() == bb.YearDay()
}

// appendBondProgress appends endgame info ONLY when user is bonded (love==100).
func (ca *CatActions) appendBondProgress(player string, msg string) string {
	if ca.LoveMeter == nil {
		return msg
	}

	// IMPORTANT: call AFTER love has been increased/decreased.
	pts, streak, err := ca.LoveMeter.RecordInteraction(context.Background(), player)
	_ = err // optionally log

	if ca.LoveMeter.Get(player) != 100 {
		return msg
	}

	if pts > 0 {
		return msg + fmt.Sprintf(" | Bonded streak: %d day(s) | +%d BondPoints", streak, pts)
	}

	return msg + fmt.Sprintf(" | Bonded streak: %d day(s) | BondPoints already earned today", streak)
}

// ExecuteAction handles player actions toward purrito
func (ca *CatActions) ExecuteAction(actionName, player, target string) string {
	t := strings.ToLower(strings.TrimSpace(target))
	a := strings.ToLower(strings.TrimSpace(actionName))

	// --- If the target is NOT purrito ---
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

	// --- If target IS purrito ---
	switch a {
	case "pet", "love":
		roll := rand.Intn(100)
		if roll < 60 {
			ca.LoveMeter.Increase(player, 1)
			return ca.acceptMessage(player)
		}

		ca.LoveMeter.Decrease(player, 1)
		return ca.rejectMessage(player)

	case "slap", "kick":
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

	case "feed":
		foods := []string{"salmon", "tuna", "sardines", "chicken", "kibble", "milk", "fish snacks", "cream", "shrimp", "turkey", "beef"}
		food := foods[rand.Intn(len(foods))]

		roll := rand.Intn(100)
		if roll < 60 {
			ca.LoveMeter.Increase(player, 1)
			return ca.feedAcceptMessage(player, food)
		}
		ca.LoveMeter.Decrease(player, 1)
		return ca.feedRejectMessage(player, food)

	case "status":
		return ca.statusMessage(player)

	case "catnip":
		return ca.catnipMessage(player)

	default:
		return "purrito tilts its head, not sure what you mean 🐾"
	}
}

// acceptMessage generates a happy response from purrito, with mood+bar (+ bond progress if bonded)
func (ca *CatActions) acceptMessage(player string) string {
	emote := emotes[rand.Intn(len(emotes))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	base := fmt.Sprintf("%s at %s and your love meter is now %d%% and purrito is now %s %s",
		emote, player, love, mood, bar)

	return ca.appendBondProgress(player, base)
}

// rejectMessage generates a grumpy response, with mood+bar (+ bond progress if still bonded)
func (ca *CatActions) rejectMessage(player string) string {
	reject := rejects[rand.Intn(len(rejects))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	base := fmt.Sprintf("purrito %s at %s and your love meter is now %d%% and purrito is now %s %s",
		reject, player, love, mood, bar)

	return ca.appendBondProgress(player, base)
}

// feedAcceptMessage – happy food reaction (+ bond progress)
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

// feedRejectMessage – picky cat reaction (+ bond progress if still bonded)
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

// statusMessage – show the player's bond with Purrito (NO bond progress here)
func (ca *CatActions) statusMessage(player string) string {
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	return fmt.Sprintf("Purrito status for %s and your love meter is %d%% and purrito is now %s %s",
		player, love, mood, bar)
}

// catnipMessage handles !catnip purrito logic (once per day, 70% accept / 30% reject)
func (ca *CatActions) catnipMessage(player string) string {
	key := strings.ToLower(strings.TrimSpace(player))
	now := time.Now()

	ca.mu.Lock()
	last, used := ca.catnipUsedAt[key]
	if used && sameDayNY(last, now) {
		ca.mu.Unlock()
		return fmt.Sprintf("aww %s, you already used catnip today. Try again tomorrow...", player)
	}
	ca.catnipUsedAt[key] = now
	ca.mu.Unlock()

	roll := rand.Intn(100)

	if roll < 70 {
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

// GetActions returns all available cat actions
func (ca *CatActions) GetActions() []string { return ca.Actions }

// GetRandomAction picks a random action from the list
func (ca *CatActions) GetRandomAction() string {
	return ca.Actions[rand.Intn(len(ca.Actions))]
}
