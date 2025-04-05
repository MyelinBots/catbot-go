package cat_actions

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
)

type CatActionsImpl interface {
	GetActions() []string
	GetRandomAction() string
	ExecuteAction(actionName, player, target string) string
}

type CatActions struct {
	LoveMeter lovemeter.LoveMeter
	Actions   []string
}

var emotes = []string{
	"meows happily",
	"rubs against your leg",
	"purrs warmly",
	"nuzzles you gently",
	"flicks its tail playfully",
}

func NewCatActions() CatActionsImpl {
	return &CatActions{
		LoveMeter: lovemeter.NewLoveMeter(),
		Actions:   emotes,
	}
}

func (ca *CatActions) ExecuteAction(actionName, player, target string) string {
	if strings.ToLower(target) != "purrito" {
		return fmt.Sprintf("%s, you can only interact with purrito.", player)
	}

	switch actionName {
	case "pet":
		ca.LoveMeter.Increase(player, 1)
		return ca.reactionMessage(player)

	case "kick":
		ca.LoveMeter.Decrease(player, 15)
		return fmt.Sprintf("purrito hisses and hides from %s! (Love: %d%%) %s", player, ca.LoveMeter.Get(player), ca.LoveMeter.GetLoveBar(player))

	default:
		return "purrito doesn't understand what you're doing."
	}
}

func (ca *CatActions) reactionMessage(player string) string {
	rand.Seed(time.Now().UnixNano())
	emote := emotes[rand.Intn(len(emotes))]
	love := ca.LoveMeter.Get(player)
	return fmt.Sprintf("%s at %s and your love meter is now %d%% 😽 %s", emote, player, love, ca.LoveMeter.GetLoveBar(player))
}

// GetActions returns the list of actions
func (ca *CatActions) GetActions() []string {
	return ca.Actions
}

// GetRandomAction returns a random action from RandomActions
func (ca *CatActions) GetRandomAction() string {
	rand.Seed(time.Now().UnixNano())
	return ca.Actions[rand.Intn(len(ca.Actions))]
}
