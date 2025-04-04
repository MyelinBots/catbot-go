package cat_actions

import (
	"fmt"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
)

type Action struct {
	Name     string
	Response string
}

func (a Action) Respond() string {
	return a.Response
}

type CatActions struct {
	LoveMeter lovemeter.LoveMeter
	Actions   map[string]Action
}

func NewCatActions(lm lovemeter.LoveMeter) *CatActions {
	return &CatActions{
		LoveMeter: lm,
		Actions: map[string]Action{
			"pet": {Name: "pet", Response: "You pet purrito!"},
		},
	}
}

// ExecuteAction executes the given action (currently supports only "pet")
func (ca *CatActions) ExecuteAction(actionName string, extras ...string) string {
	action, exists := ca.Actions[actionName]
	if !exists {
		return "Unknown action."
	}

	// Extract target name (optional)
	var target string
	if len(extras) > 0 {
		target = strings.ToLower(extras[0])
	}

	if actionName == "pet" {
		if target != "purrito" {
			return "You can only pet purrito."
		}

		if ca.LoveMeter.Get() < 70 {
			return "purrito doesn't want to be petted right now. 😾"
		}

		ca.LoveMeter.Increase(5)
		return fmt.Sprintf("%s Love meter: %d", action.Respond(), ca.LoveMeter.Get())
	}

	return "Unknown action."
}
