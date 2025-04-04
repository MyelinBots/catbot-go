package cat_actions

import (
	"fmt"

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

func (ca *CatActions) ExecuteAction(actionName string) string {
	action, exists := ca.Actions[actionName]
	if !exists {
		return "Unknown action."
	}

	if actionName == "pet" {
		if ca.LoveMeter.Get() < 70 {
			return "purrito doesn't want to be petted right now."
		}
		ca.LoveMeter.Increase(5)
	}

	return fmt.Sprintf("%s Love meter: %d", action.Respond(), ca.LoveMeter.Get())
}
