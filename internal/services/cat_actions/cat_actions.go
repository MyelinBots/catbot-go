package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// CatActions represents the cat's actions and state
type CatActions struct {
	Actions       []string
	RandomActions []string
	LoveMeter     int
}

// NewCatActions creates a new CatActions instance
func NewCatActions() *CatActions {
	return &CatActions{
		Actions: []string{
			"purr", "meow", "hiding", "hiss", "jump", "play", "scratch", "rub against", "surprise",
		},
		RandomActions: []string{"meow", "purr", "rub against"},
		LoveMeter:     0,
	}
}

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

func (ca *CatActions) GetLoveMeter() int {
	return ca.LoveMeter
}

// GetActions returns the list of actions
func (ca *CatActions) GetActions() []string {
	return ca.Actions
}

// GetRandomAction returns a random action from RandomActions
func (ca *CatActions) GetRandomAction() string {
	rand.Seed(time.Now().UnixNano())
	return ca.RandomActions[rand.Intn(len(ca.RandomActions))]
}

// IncreaseLoveMeter increases the love meter by 1
func (ca *CatActions) IncreaseLoveMeter() {
	ca.LoveMeter++
}

// Action represents a specific action with logic to process a message
type Action struct {
	Type   string
	Nick   string
	Needle string
}

// NewAction creates a new Action instance
func NewAction(actionType, nick, needle string) *Action {
	return &Action{
		Type:   actionType,
		Nick:   nick,
		Needle: needle,
	}
}

// CanProcess checks if the action can process the given message
func (a *Action) CanProcess(message string) bool {
	return strings.Contains(strings.ToLower(message), strings.ToLower(a.Needle))
}

// Process handles the action logic based on the message and updates the cat state
func (a *Action) Process(message string, cat *CatActions) {
	message = strings.ToLower(message)
	parts := strings.Split(message, a.Needle)

	if len(parts) < 2 {
		fmt.Println("No specific action to process.")
		return
	}

	part := strings.TrimSpace(parts[1])
	switch {
	case strings.Contains(part, "head"):
		fmt.Println("Petting the head")
		cat.IncreaseLoveMeter()
	case strings.Contains(part, "tail"):
		fmt.Println("Petting the tail")
		cat.IncreaseLoveMeter()
	case strings.Contains(part, "body"):
		fmt.Println("Petting the body")
		cat.IncreaseLoveMeter()
	default:
		fmt.Println("Petting the cat")
		cat.IncreaseLoveMeter()
	}

	// Cat response after petting
	fmt.Printf("Cat response: %s\n", cat.GetRandomAction())
}

func main() {
	nick := "purrito"
	petAction := NewAction("pet", nick, "pet "+nick)

	// Simulating a user command
	message := "pet purrito head" // Simulated user command

	cat := NewCatActions()

	if petAction.CanProcess(message) {
		petAction.Process(message, cat)
	} else {
		fmt.Println("No action can be processed.")
	}
}
