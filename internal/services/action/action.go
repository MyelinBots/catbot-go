package main

import (
	"fmt"
	"strings"
)

// CatActions represents the cat's actions and its love meter
type CatActions struct {
	LoveMeter int
}

// NewCatActions creates a new CatActions instance
func NewCatActions() *CatActions {
	return &CatActions{
		LoveMeter: 0,
	}
}

// AdjustLoveMeter adjusts the love meter based on the action
func (ca *CatActions) AdjustLoveMeter(action string) {
	switch action {
	case "pet", "feed":
		ca.LoveMeter += 10
		if ca.LoveMeter > 100 {
			ca.LoveMeter = 100
		}
		fmt.Println("The cat purrs happily!")
	case "kick":
		ca.LoveMeter -= 10
		if ca.LoveMeter < 0 {
			ca.LoveMeter = 0
		}
		fmt.Println("The cat hisses angrily!")
	default:
		fmt.Println("The cat is indifferent.")
	}
}

// GetLoveMeter returns the current love meter value
func (ca *CatActions) GetLoveMeter() int {
	return ca.LoveMeter
}

// Action represents a specific action
type Action struct {
	Type     string
	Nick     string
	Response string
}

// NewAction creates a new Action instance
func NewAction(actionType, nick, response string) *Action {
	return &Action{
		Type:     actionType,
		Nick:     nick,
		Response: response,
	}
}

// String returns the type of the action
func (a *Action) String() string {
	return a.Type
}

// Respond returns the response of the action
func (a *Action) Respond(player string) string {
	return fmt.Sprintf("%s: %s", player, a.Response)
}

// Process processes the action and updates the cat's love meter
func (a *Action) Process(message string, cat *CatActions) {
	message = strings.ToLower(message)
	fmt.Printf("Processing action '%s' for '%s'\n", a.Type, a.Nick)

	// Adjust the love meter based on the action
	switch {
	case strings.Contains(message, "pet"):
		fmt.Println("Petting the cat...")
		cat.AdjustLoveMeter("pet")
	case strings.Contains(message, "feed"):
		fmt.Println("Feeding the cat...")
		cat.AdjustLoveMeter("feed")
	case strings.Contains(message, "kick"):
		fmt.Println("Kicking the cat...")
		cat.AdjustLoveMeter("kick")
	default:
		fmt.Println("Unknown action.")
	}

	// Display the current love meter count
	fmt.Printf("Love Meter: %d\n", cat.GetLoveMeter())
}

func main() {
	// Initialize the cat actions
	cat := NewCatActions()

	// Example usage
	action := NewAction("feed", "purrito", "You fed the cat!")
	fmt.Println(action.String())          // Output: feed
	fmt.Println(action.Respond("Player")) // Output: Player: You fed the cat!

	// Process commands
	commands := []string{"pet purrito", "feed purrito", "kick purrito", "pet purrito"}
	for _, message := range commands {
		action.Process(message, cat)
	}
}
