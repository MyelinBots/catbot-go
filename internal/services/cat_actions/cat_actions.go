package cat_actions

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
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
}

var emotes = []string{
	"meows happily",
	"rubs against your leg",
	"purrs warmly",
	"nuzzles you gently",
	"flicks its tail playfully",
	"stretches and yawns",
	"rolls over for belly rubs",
	"gives a soft chirp",
	"licks its paw and looks",
	"blinks slowly",
	"purrs contentedly",
}

// NewCatActions returns a new instance of CatActions
func NewCatActions(catPlayerRepo cat_player.CatPlayerRepository, network string, channel string) CatActionsImpl {
	return &CatActions{
		LoveMeter:     lovemeter.NewLoveMeter(catPlayerRepo, network, channel),
		Actions:       emotes,
		CatPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,
	}
}

// ExecuteAction handles player actions toward purrito
func (ca *CatActions) ExecuteAction(actionName, player, target string) string {
	if strings.ToLower(target) != "purrito" {
		return fmt.Sprintf("%s, you can only interact with purrito.", player)
	}

	switch actionName {
	case "pet":
		love := ca.LoveMeter.Get(player)

		// Higher rejection chance if love is low
		rejectChance := 5
		if love < 20 {
			rejectChance = 3
		}

		if rand.Intn(rejectChance) == 0 {
			ca.LoveMeter.Decrease(player, 5)
			rejects := []string{
				"hisses and moves away",
				"growls softly, not in the mood",
				"glares coldly",
				"turns their back",
				"gives disdainful look",
				"flicks its tail in annoyance",
				"lets out a displeased meow",
				"stiffens and walks away",
				"gives a sharp meow and walks off",
				"scratches the ground and ignores you",
				"gives a dismissive flick of the tail",
			}
			rejectMsg := rejects[rand.Intn(len(rejects))]
			return fmt.Sprintf("purrito %s at %s and your love meter decreased to %d%% 😾 %s",
				rejectMsg, player, ca.LoveMeter.Get(player), ca.LoveMeter.GetLoveBar(player))
		}

		// Accepted petting
		ca.LoveMeter.Increase(player, 1)
		return ca.reactionMessage(player)

	default:
		return "purrito doesn't understand what you're doing."
	}
}

// reactionMessage generates a happy response from purrito
func (ca *CatActions) reactionMessage(player string) string {
	rand.Seed(time.Now().UnixNano())
	emote := emotes[rand.Intn(len(emotes))]
	love := ca.LoveMeter.Get(player)
	return fmt.Sprintf("%s at %s and your love meter is now %d%% 😽 %s",
		emote, player, love, ca.LoveMeter.GetLoveBar(player))
}

// GetActions returns all available cat actions
func (ca *CatActions) GetActions() []string {
	return ca.Actions
}

// GetRandomAction picks a random action from the list
func (ca *CatActions) GetRandomAction() string {
	rand.Seed(time.Now().UnixNano())
	return ca.Actions[rand.Intn(len(ca.Actions))]
}
