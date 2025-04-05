package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// CatActions represents the cat's actions and its love meter
type CatActions struct {
	LoveMeter int
	Actions   map[string]Action
}

// Action represents a specific action
type Action struct {
	Type     string
	Response string
}

func (a Action) Respond(nick string, love int) string {
	emotes := []string{
		"meows happily", "rubs against your leg", "purrs warmly", "nuzzles you gently", "flicks its tail playfully",
	}
	rand.Seed(time.Now().UnixNano())
	reaction := emotes[rand.Intn(len(emotes))]
	return fmt.Sprintf("%s at %s and your love meter increased to %d%% ❤️", reaction, nick, love)
}

func NewCatActions() *CatActions {
	return &CatActions{
		LoveMeter: 0,
		Actions: map[string]Action{
			"pet":  {Type: "pet", Response: "You pet the cat!"},
			"kick": {Type: "kick", Response: "You kicked the cat!"},
		},
	}
}

func (ca *CatActions) ExecuteAction(actionName, nick, target string) string {
	action, exists := ca.Actions[actionName]
	if !exists {
		return "Unknown action."
	}

	if strings.ToLower(target) != "purrito" {
		return "You can only interact with purrito."
	}

	switch actionName {
	case "pet":
		if ca.LoveMeter < 70 {
			ca.LoveMeter += 10
			return fmt.Sprintf("looks at %s cautiously... but doesn't run away. (Love: %d%%)", nick, ca.LoveMeter)
		}
		ca.LoveMeter += 10
		if ca.LoveMeter > 100 {
			ca.LoveMeter = 100
		}
		return action.Respond(nick, ca.LoveMeter)

	case "kick":
		ca.LoveMeter -= 15
		if ca.LoveMeter < 0 {
			ca.LoveMeter = 0
		}
		return fmt.Sprintf("hisses and hides from %s! (Love: %d%%)", nick, ca.LoveMeter)

	default:
		return "purrito doesn't understand what you're doing."
	}
}
