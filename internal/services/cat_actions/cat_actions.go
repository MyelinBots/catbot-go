package cat_actions

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
)

// Seed RNG once for the package
func init() { rand.Seed(time.Now().UnixNano()) }

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
}

var emotes = []string{
	"meows happily",
	"rubs against leg",
	"purrs warmly",
	"nuzzles gently",
	"flicks its tail playfully",
	"stretches and yawns",
	"rolls over for belly rubs",
	"gives a soft chirp",
	"licks its paw and looks",
	"blinks slowly",
	"purrs contentedly",
	"curls up beside you",
	"gives a gentle headbutt",
	"flicks its ears",
	"swishes its tail",
	"paws at the air",
	"gives a playful swipe",
	"chases a sunbeam",
	"sniffs curiously",
	"gives a happy meow",
	"pounces playfully",
	"gives a soft trill",
}

var rejects = []string{
	"hisses and moves away",
	"growls softly, not in the mood",
	"glares coldly",
	"turns their back",
	"gives a disdainful look",
	"flicks its tail in annoyance",
	"lets out a displeased meow",
	"stiffens and walks away",
	"gives a sharp meow and walks off",
	"scratches the ground and ignores you",
	"gives a dismissive flick of the tail",
	"ears flatten in irritation",
	"gives a warning hiss",
	"swats the air and moves away",
	"gives a disdainful glance",
	"turns its head away",
	"gives a sharp meow and walks off",
	"ignores you completely",
	"gives a cold stare",
	"flicks its tail and walks away",
	"lets out an annoyed meow",
}

// NewCatActions returns a new instance of CatActions
func NewCatActions(catPlayerRepo cat_player.CatPlayerRepository, network, channel string) CatActionsImpl {
	return &CatActions{
		LoveMeter:     lovemeter.NewLoveMeter(catPlayerRepo, network, channel),
		Actions:       emotes,
		CatPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,
	}
}

// ExecuteAction handles player actions toward purrito
// 60% chance to ACCEPT a pet, 40% to REJECT.
func (ca *CatActions) ExecuteAction(actionName, player, target string) string {
	if strings.ToLower(strings.TrimSpace(target)) != "purrito" {
		return fmt.Sprintf("%s, you can only interact with purrito.", player)
	}

	switch strings.ToLower(strings.TrimSpace(actionName)) {
	case "pet", "love":
		roll := rand.Intn(100) // 0..99
		if roll < 60 {
			// ACCEPT (increase by 1)
			ca.LoveMeter.Increase(player, 1)
			return ca.acceptMessage(player)
		}
		// REJECT (decrease by 1)
		ca.LoveMeter.Decrease(player, 1)
		return ca.rejectMessage(player)

	default:
		return "purrito tilts its head, not sure what you mean 🐾"
	}
}

// acceptMessage generates a happy response from purrito, with mood+bar
func (ca *CatActions) acceptMessage(player string) string {
	emote := emotes[rand.Intn(len(emotes))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)
	return fmt.Sprintf("%s at %s and your love meter is now %d%% and purrito is now %s %s",
		emote, player, love, mood, bar)
}

// rejectMessage generates a grumpy response, with mood+bar
func (ca *CatActions) rejectMessage(player string) string {
	reject := rejects[rand.Intn(len(rejects))]
	love := ca.LoveMeter.Get(player)
	mood := ca.LoveMeter.GetMood(player)
	bar := ca.LoveMeter.GetLoveBar(player)
	return fmt.Sprintf("purrito %s at %s and your love meter is now %d%% and purrito is now %s %s",
		reject, player, love, mood, bar)
}

// GetActions returns all available cat actions
func (ca *CatActions) GetActions() []string { return ca.Actions }

// GetRandomAction picks a random action from the list
func (ca *CatActions) GetRandomAction() string {
	return ca.Actions[rand.Intn(len(ca.Actions))]
}
