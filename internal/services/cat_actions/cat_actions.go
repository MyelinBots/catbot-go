package cat_actions

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

	// spawn session
	presentUntil time.Time
	nextSpawnAt  time.Time

	minRespawn  time.Duration
	maxRespawn  time.Duration
	spawnWindow time.Duration

	lastLeaveMsg string
	lastSpawnMsg string
}

func NewCatActions(catPlayerRepo cat_player.CatPlayerRepository, network, channel string, spawnWindow, minRespawn, maxRespawn time.Duration) CatActionsImpl {
	ca := &CatActions{
		LoveMeter:     lovemeter.NewLoveMeter(catPlayerRepo, network, channel),
		BondPoints:    bondpoints.New(catPlayerRepo),
		Actions:       emotes,
		CatPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,

		slapWarned:   make(map[string]bool),
		catnipUsedAt: make(map[string]time.Time),

		minRespawn:  minRespawn,
		maxRespawn:  maxRespawn,
		spawnWindow: spawnWindow,
	}

	// Start present immediately
	now := time.Now()
	ca.presentUntil = now.Add(ca.spawnWindow)

	return ca
}

// --------------------
// Basic
// --------------------

func (ca *CatActions) GetActions() []string { return ca.Actions }

func (ca *CatActions) GetRandomAction() string {
	return ca.Actions[rand.Intn(len(ca.Actions))]
}

func (ca *CatActions) appendBondProgress(_ string, msg string) string { return msg }

// --------------------
// Helpers
// --------------------

func normalizeAction(action string) string {
	a := strings.ToLower(strings.TrimSpace(action))
	return strings.TrimPrefix(a, "!")
}

func formatWait(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	sec := int(d.Seconds())
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	min := sec / 60
	sec = sec % 60
	if min < 60 {
		return fmt.Sprintf("%dm %ds", min, sec)
	}
	hr := min / 60
	min = min % 60
	return fmt.Sprintf("%dh %dm", hr, min)
}

func sameDayInNY(a, b time.Time) bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}

	aa := a.In(loc)
	bb := b.In(loc)

	return aa.Year() == bb.Year() && aa.YearDay() == bb.YearDay()
}

func giftNamesFromMask(mask int) []string {
	var out []string

	if mask&bondrewards.Gift7 != 0 {
		out = append(out, "🐹 Tiny Guinea Pig")
	}
	if mask&bondrewards.Gift14 != 0 {
		out = append(out, "🐍 Cute Python")
	}
	if mask&bondrewards.Gift21 != 0 {
		out = append(out, "🦜 Noisy Parrot")
	}
	if mask&bondrewards.Gift30 != 0 {
		out = append(out, "🐠 Colorful Fish")
	}
	if mask&bondrewards.Gift45 != 0 {
		out = append(out, "🐱 Friendly Kitten")
	}
	if mask&bondrewards.Gift100 != 0 {
		out = append(out, "🎁 Secret Gift (Forever Human)")
	}

	return out
}

func misuseMessage(player, action, target string) string {
	pt := cases.Title(language.English).String(target)

	// Special: slap misuse => cat retaliates
	if action == "slap" {
		lines := []string{
			fmt.Sprintf("😾 scratches %s's face hardly... Why did you slap %s?", player, pt),
			fmt.Sprintf("😼 hisses and swats %s... Don’t slap %s", player, pt),
			fmt.Sprintf("🐾 claws %s... I did not like you slapping %s", player, pt),
			fmt.Sprintf("😿 bites %s lightly... Why would you slap %s?", player, pt),
		}
		return lines[rand.Intn(len(lines))]
	}

	// Optional: kick misuse => cat retaliates
	if action == "kick" {
		lines := []string{
			fmt.Sprintf("😾 lunges at %s... Don’t kick %s!", player, pt),
			fmt.Sprintf("🐾 scratches %s’s leg... Kicking %s is not okay :(", player, pt),
			fmt.Sprintf("😼 hisses at %s... Why would you kick %s?", player, pt),
			fmt.Sprintf("😿 bites %s’s ankle... Kicking %s made me sad...", player, pt),
		}
		return lines[rand.Intn(len(lines))]
	}

	if action == "laser" {
		lines := []string{
			fmt.Sprintf("😼 Why are you using the laser on %s? ... you have to -> !%s purrito", pt, action),
			fmt.Sprintf("😼 %s is not purrito.,, You have to -> !%s purrito", pt, action),
			fmt.Sprintf("🐾 Wrong target, %s... You have to -> !%s purrito", player, action),
			fmt.Sprintf("😿 You seem confused... You have to -> !%s purrito", action),
		}
		return lines[rand.Intn(len(lines))]
	}

	// Generic misuse for all other commands
	verb := action + "ing"
	switch action {
	case "pet":
		verb = "petting"
	case "love":
		verb = "loving"
	case "feed":
		verb = "feeding"
	case "laser":
		verb = "using the laser on"
	case "catnip":
		verb = "giving catnip to"
	}

	lines := []string{
		fmt.Sprintf("😼 %s blinks at you... You are %s %s, but %s is not me.", pt, verb, pt, pt),
		fmt.Sprintf("🐾 %s tilts its head in confusion... Why are you %s %s?", pt, verb, pt),
		fmt.Sprintf("😿 %s looks awkward... I think you meant to do that to Purrito.", pt),
		fmt.Sprintf("😼 %s ignores you completely... That command is not for me.", pt),
	}
	return lines[rand.Intn(len(lines))]
}

// --------------------
// Spawn / Presence
// --------------------

func (ca *CatActions) IsHere() bool {
	now := time.Now()

	ca.mu.Lock()
	defer ca.mu.Unlock()

	// present expired => despawn + schedule respawn (timeout leave)
	if !ca.presentUntil.IsZero() && now.After(ca.presentUntil) {
		ca.lastLeaveMsg = timeoutLeaveMessage() // ✅ now it's used
		ca.despawnLocked(now)
	}

	// not present but respawn time reached => spawn again
	if ca.presentUntil.IsZero() && !ca.nextSpawnAt.IsZero() && !now.Before(ca.nextSpawnAt) {
		ca.presentUntil = now.Add(ca.spawnWindow)

		// ✅ ตั้งข้อความ "โผล่" แค่ครั้งเดียวต่อรอบ
		emote := emotes[rand.Intn(len(emotes))]
		ca.lastSpawnMsg = fmt.Sprintf("🐈 meowww ... %s", emote)
	}

	return !ca.presentUntil.IsZero() && now.Before(ca.presentUntil)
}

// EnsureHere is kept for backward compatibility (catbot.go still calls it).
// With the "one interaction per spawn" system, we DO NOT want callers to keep
// Purrito permanently present.
//
// EnsureHere only spawns Purrito if he is not present AND there is no pending spawn timer.
// If he is present, it does nothing (does NOT extend the window).
func (ca *CatActions) EnsureHere(forHowLong time.Duration) {
	now := time.Now()

	ca.mu.Lock()
	defer ca.mu.Unlock()

	// If already present, do not extend (prevents "always here" behavior)
	if !ca.presentUntil.IsZero() && now.Before(ca.presentUntil) {
		return
	}

	// If there is a future spawn scheduled, respect it
	if !ca.nextSpawnAt.IsZero() && now.Before(ca.nextSpawnAt) {
		return
	}

	// Spawn now for at most forHowLong, but cap to ca.spawnWindow to keep gameplay consistent
	window := forHowLong
	if window <= 0 || window > ca.spawnWindow {
		window = ca.spawnWindow
	}

	ca.presentUntil = now.Add(window)
	ca.nextSpawnAt = time.Time{}
}

func (ca *CatActions) despawnLocked(now time.Time) {
	ca.presentUntil = time.Time{}

	delay := ca.minRespawn
	if ca.maxRespawn > ca.minRespawn {
		delay = ca.minRespawn + time.Duration(rand.Int63n(int64(ca.maxRespawn-ca.minRespawn)))
	}
	ca.nextSpawnAt = now.Add(delay)
}

// DespawnAfterInteraction immediately despawns Purrito and starts the respawn timer.
// Call this after a successful interaction to enforce "one interaction per spawn".
func (ca *CatActions) DespawnAfterInteraction() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.despawnLocked(time.Now())
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

//func (ca *CatActions) gatePresenceForAction(_ string) (bool, string) {
//	if ca.IsHere() {
//		return true, ""
//	}
//
//	ca.mu.RLock()
//	next := ca.nextSpawnAt
//	ca.mu.RUnlock()
//
//	wait := time.Duration(0)
//	if !next.IsZero() {
//		wait = time.Until(next)
//	}
//
//	return false, fmt.Sprintf("🐾 Purrito is not here right now... try again in %s!", formatWait(wait))
//}
// --------------------
// Main Execute
// --------------------

func (ca *CatActions) gatePresenceForAction(_ string) (bool, string) {
	// if he just left because of timeout, show that message once
	if msg := ca.PopLeaveMessage(); msg != "" {
		return false, msg
	}

	if ca.IsHere() {
		return true, ""
	}

	ca.mu.RLock()
	next := ca.nextSpawnAt
	ca.mu.RUnlock()

	wait := time.Duration(0)
	if !next.IsZero() {
		wait = time.Until(next)
	}

	return false, fmt.Sprintf("🐾 Purrito is not here right now... he will be back in %s...", formatWait(wait))
}

func (ca *CatActions) ExecuteAction(actionName, player, target string) string {
	t := normalizeAction(target)
	a := normalizeAction(actionName)

	// status can be used without targeting purrito (optional rule)
	if a == "status" {
		return ca.statusMessage(player)
	}

	// all other commands must target purrito
	if t != "purrito" {
		return misuseMessage(player, a, target)
	}

	switch a {

	case "pet", "love":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}

		// One interaction per spawn - despawn immediately
		ca.DespawnAfterInteraction()

		if rand.Intn(100) < 60 {
			ca.LoveMeter.Increase(player, 1)
			return ca.acceptMessage(player)
		}

		ca.LoveMeter.Decrease(player, 1)
		return ca.rejectMessage(player)

	case "feed":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}

		// One interaction per spawn - despawn immediately
		ca.DespawnAfterInteraction()

		foods := []string{
			"salmon", "tuna", "sardines", "chicken", "kibble", "milk",
			"fish snacks", "cream", "shrimp", "turkey", "beef", "cat treats",
			"catnip-infused snacks",
		}
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

		// One interaction per spawn - despawn immediately
		ca.DespawnAfterInteraction()

		if rand.Intn(100) < 60 {
			ca.LoveMeter.Increase(player, 1)
			return ca.laserAcceptMessage(player)
		}

		ca.LoveMeter.Decrease(player, 1)
		return ca.laserRejectMessage(player)

	case "catnip":
		if ok, msg := ca.gatePresenceForAction(a); !ok {
			return msg
		}

		// cooldown rejection does NOT cause Purrito to leave (exception for catnip)
		if ca.CatnipOnCooldown(player) {
			rem := ca.CatnipRemaining(player)
			return fmt.Sprintf("aww %s, you already used catnip today. Try again in %s", player, formatRemaining(rem))
		}

		// Successful catnip - one interaction per spawn, despawn immediately
		ca.DespawnAfterInteraction()

		return ca.catnipMessage(player)

	case "slap", "kick":
		// slap/kick does not require presence by default (but still must target purrito)

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
				fmt.Sprintf("⚠️ Purrito stares at %s with shocked eyes... he did not like that...", player),
				fmt.Sprintf("😿 Purrito backs away from %s...please be gentle with him", player),
				fmt.Sprintf("⚠️ Purrito watches %s carefully... one more slap and he will be upset", player),
				fmt.Sprintf("😼 Purrito lifts a paw at %s in warning... do not try that again...", player),
			}
			return firstWarnings[rand.Intn(len(firstWarnings))]
		}

		ca.LoveMeter.Decrease(player, 1)
		love := ca.LoveMeter.Get(player)
		mood := ca.LoveMeter.GetMood(player)
		bar := ca.LoveMeter.GetLoveBar(player)

		secondPunishments := []string{
			fmt.Sprintf("😾 Purrito swats back at %s and looks hurt. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😾 Purrito hisses softly at %s... his heart hurts. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😿 Purrito lowers his ears... %s made him sad. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😿 Purrito looks betrayed by %s. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
			fmt.Sprintf("😾 Purrito steps back from %s... do not hurt him. your love meter decreased to %d%% and purrito is now %s %s", player, love, mood, bar),
		}
		return secondPunishments[rand.Intn(len(secondPunishments))]

	default:
		return "purrito tilts its head, don't know what you mean 🐾"
	}
}

// --------------------
// Messages
// --------------------

func (ca *CatActions) acceptMessage(player string) string {
	emote := emotes[rand.Intn(len(emotes))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	bonus := ca.tryAwardBondPoints(player) // ✅ เพิ่ม

	base := fmt.Sprintf("%s at %s and your love meter is now %d%% and purrito is now %s %s%s",
		emote, player, love, mood, bar, bonus)

	return base
}

func (ca *CatActions) rejectMessage(player string) string {
	reject := rejects[rand.Intn(len(rejects))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	base := fmt.Sprintf("purrito %s at %s and your love meter is now %d%% and purrito is now %s %s", reject, player, love, mood, bar)
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
		fmt.Sprintf("😸 Purrito licks his lips after eating the %s from %s! Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
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
		fmt.Sprintf("😿 Purrito walks away from the %s offered by %s... Your love meter is now %d%% and purrito is now %s %s", food, player, love, mood, bar),
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
	// LoveMeter / Mood
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)

	// Presence
	isHere := ca.IsHere()

	ca.mu.RLock()
	presentUntil := ca.presentUntil
	nextSpawn := ca.nextSpawnAt
	ca.mu.RUnlock()

	now := time.Now()

	// --- Presence line (colored) ---
	var presenceLine string
	if isHere && !presentUntil.IsZero() {
		presenceLine = fmt.Sprintf(
			"\x0310🐾 Presence:\x0F \x0303HERE\x0F (leaves in %s)",
			formatWait(time.Until(presentUntil)),
		)
	} else {
		wait := time.Duration(0)
		if !nextSpawn.IsZero() && now.Before(nextSpawn) {
			wait = time.Until(nextSpawn)
		}
		presenceLine = fmt.Sprintf(
			"\x0310🐾 Presence:\x0F \x0304AWAY\x0F (back in %s)",
			formatWait(wait),
		)
	}

	// --- Catnip cooldown (colored) ---
	rem := ca.CatnipRemaining(player)
	catnipLine := "\x0310🌿 Catnip:\x0F \x0303READY\x0F"
	if rem > 0 {
		catnipLine = fmt.Sprintf(
			"\x0310🌿 Catnip:\x0F \x0308USED\x0F (%s left)",
			formatWait(rem),
		)
	}

	// --- Load player record (read-only) ---
	p, err := ca.CatPlayerRepo.GetPlayerByName(context.Background(), player, ca.Network, ca.Channel)
	if err != nil || p == nil {
		lines := []string{
			fmt.Sprintf("\x0310😺 Purrito Status for:\x0F \x0300%s\x0F", player),
			fmt.Sprintf("\x0310Love meter:\x0F %d%%  \x0310Mood:\x0F %s %s", love, mood, bar),
			presenceLine,
			catnipLine,
		}
		return strings.Join(lines, " | ")
	}

	// --- Main progression (HighestStreak + title) ---
	title := bondrewards.TitleForHighestStreak(p.HighestStreak)
	mainLine := fmt.Sprintf("\x0310HighestStreak:\x0F %d | \x0310Title:\x0F %s", p.HighestStreak, title)

	// --- BondPoints progression ---
	bpLine := fmt.Sprintf(
		"\x0310BondPoints:\x0F %d | \x0310Streak:\x0F %d | \x0310HighestBP:\x0F %d",
		p.BondPoints, p.BondPointStreak, p.HighestBondStreak,
	)

	// --- BondPoints today availability (colored) ---
	bpReady := "\x0304LOCKED\x0F (need HighestStreak \u2265 100)"
	if p.HighestStreak >= 100 {
		bpReady = "\x0303READY\x0F"
		if p.LastBondPointsAt != nil && sameDayInNY(*p.LastBondPointsAt, now) {
			bpReady = "\x0308ALREADY AWARDED TODAY\x0F"
		}
	}
	bpReadyLine := fmt.Sprintf("\x0310BondPoints Today:\x0F %s", bpReady)

	// --- Gifts (colored label) ---
	gifts := giftNamesFromMask(p.GiftsUnlocked)
	giftsLine := fmt.Sprintf("\x0310Gifts:\x0F %s", bondrewards.JoinGifts(gifts))

	// --- Final output (multi-line, readable) ---
	lines := []string{
		fmt.Sprintf("\x0310😺 Purrito Status for:\x0F \x0300%s\x0F", player),
		fmt.Sprintf("\x0310Love meter:\x0F %d%%  \x0310Mood:\x0F %s %s", love, mood, bar),
		presenceLine,
		catnipLine,
		mainLine,
		bpLine,
		bpReadyLine,
		giftsLine,
	}

	return strings.Join(lines, " | ")
}

// catnipMessage assumes cooldown was checked BEFORE calling it.
func (ca *CatActions) catnipMessage(player string) string {
	key := normalizeNick(player)
	now := time.Now()

	ca.mu.Lock()
	ca.catnipUsedAt[key] = now
	ca.mu.Unlock()

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

// timeoutLeaveMessage returns a message when Purrito leaves because he stayed
// for the full spawnWindow (timeout) and nobody interacted.
func timeoutLeaveMessage() string {
	lines := []string{
		"(=^‥^=)っ ...looks around... no one came. He quietly walks away...",
		"(=^‥^=)っ ...stretches, yawns, and wanders off...",
		"(=^‥^=)っ ...blinks slowly... then disappears into the night...",
		"(=^‥^=)っ ...waits patiently... then gives up and leaves...",
		"(=^‥^=)っ ...decides to explore somewhere else and slips away...",
		"(=^‥^=)っ ...flicks his tail, bored of waiting, and walks off...",
		"(=^‥^=)っ ...pads away softly... you barely notice he is gone...",
		"(=^‥^=)っ ...hops onto a fence and vanishes...",
		"(=^‥^=)っ ...Time’s up... Purrito got tired of waiting and left...",
	}
	return lines[rand.Intn(len(lines))]
}

// PopLeaveMessage returns the timeout leave message once (then clears it).
// ForceAbsent forces Purrito to be absent immediately, clearing any presence
// and pending spawn timers.
func (ca *CatActions) ForceAbsent() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.presentUntil = time.Time{}
	ca.nextSpawnAt = time.Time{}
}

func (ca *CatActions) PopLeaveMessage() string {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	msg := ca.lastLeaveMsg
	ca.lastLeaveMsg = ""
	return msg
}

func (ca *CatActions) PopSpawnMessage() string {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	msg := ca.lastSpawnMsg
	ca.lastSpawnMsg = ""
	return msg
}

func (ca *CatActions) TickPresence() (spawnMsg string, leaveMsg string) {
	// อัปเดตสถานะ (จะทำให้เกิด spawn/despawn ตามเวลา)
	_ = ca.IsHere()

	// ถ้ามี spawn/leave message ใหม่ ให้ดึงออกมา (ครั้งเดียว)
	spawnMsg = ca.PopSpawnMessage()
	leaveMsg = ca.PopLeaveMessage()
	return
}

func (ca *CatActions) tryAwardBondPoints(player string) string {
	// เรียกหลัง "สำเร็จ" เท่านั้น
	res, err := ca.BondPoints.RecordBondedInteraction(context.Background(), player, ca.Network, ca.Channel)
	if err != nil {
		// อย่าให้พังเกมหลัก แค่แนบข้อความเบาๆ
		return ""
	}

	// ถ้าวันนี้ได้ 0 แปลว่า: (1) ให้ไปแล้ววันนี้ หรือ (2) ยังไม่ถึง gate
	if res.AwardedPoints <= 0 {
		return ""
	}

	// แนบข้อความสั้นๆ ให้รู้สึก rewarding
	return fmt.Sprintf(" ✨ +%d BondPoints (Total: %d ::: BP Streak: %d)", res.AwardedPoints, res.TotalPoints, res.Streak)
}

func (ca *CatActions) HandleStatus(sender string, args []string) string {
	target := sender // default: self

	if len(args) >= 1 {
		target = args[0]
	}

	return ca.statusMessage(target)
}
