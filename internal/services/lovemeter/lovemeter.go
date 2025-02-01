package lovemeter

import (
	"fmt"
	"math/rand"
	"time"
)

type LoveMeter interface {
	Increase(amount int)
	Decrease(amount int)
}

type LoveMeterImpl struct {
	Value int
}

func NewLoveMeter() LoveMeter {
	return &LoveMeterImpl{Value: 50}
}

func (lm *LoveMeterImpl) Increase(amount int) {
	lm.Value += amount
	if lm.Value > 100 {
		lm.Value = 100
	}
}

func (lm *LoveMeterImpl) Decrease(amount int) {
	lm.Value -= amount
	if lm.Value < 0 {
		lm.Value = 0
	}
}

type Action struct {
	Name     string
	Response string
}

func (a Action) Respond() string {
	return a.Response
}

type CatActions struct {
	LoveMeter *LoveMeterImpl
	Actions   map[string]Action
}

func NewCatActions(loveMeter *LoveMeterImpl) *CatActions {
	return &CatActions{
		LoveMeter: loveMeter,
		Actions: map[string]Action{
			"feed": {Name: "feed", Response: "You fed purrito!"},
			"pet":  {Name: "pet", Response: "You pet purrito!"},
			"hug":  {Name: "hug", Response: "You hugged purrito!"},
			"kick": {Name: "kick", Response: "You kicked purrito!"},
		},
	}
}

func (ca *CatActions) ExecuteAction(actionName string) string {
	action, exists := ca.Actions[actionName]
	if !exists {
		return "Unknown action."
	}

	switch actionName {
	case "feed":
		ca.LoveMeter.Increase(10)
	case "pet":
		if ca.LoveMeter.Value < 70 {
			return "purrito doesn't want to be petted right now."
		}
		ca.LoveMeter.Increase(5)
	case "hug":
		if ca.LoveMeter.Value < 80 {
			return "purrito isn't ready for a hug."
		}
		ca.LoveMeter.Increase(10)
	case "kick":
		ca.LoveMeter.Decrease(15)
	}

	return fmt.Sprintf("%s Love meter: %d", action.Respond(), ca.LoveMeter.Value)
}

func handleCommand(command string, cat *CatActions) string {
	commandMapping := map[string]string{
		"!feed purrito": "feed",
		"!pet purrito":  "pet",
		"!hug purrito":  "hug",
		"!kick purrito": "kick",
	}

	actionName, exists := commandMapping[command]
	if !exists {
		return "Unknown command."
	}

	return cat.ExecuteAction(actionName)
}

func simulateIRCMessage(nick, message string, cat *CatActions) {
	fmt.Printf("%s: %s\n", nick, message)
	response := handleCommand(message, cat)
	fmt.Printf("Bot: %s\n", response)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	loveMeter := &LoveMeterImpl{Value: 50}
	cat := NewCatActions(loveMeter)

	// Simulating messages from users
	simulateIRCMessage("<random_nick>", "!feed purrito", cat)
	simulateIRCMessage("<random_nick>", "!pet purrito", cat)
	simulateIRCMessage("<random_nick>", "!hug purrito", cat)
	simulateIRCMessage("<random_nick>", "!kick purrito", cat)
}
