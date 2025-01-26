package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// LoveMeter represents the love meter
type LoveMeter struct {
	Value int
	mu    sync.Mutex
}

func NewLoveMeter() *LoveMeter {
	return &LoveMeter{Value: 50}
}

func (lm *LoveMeter) Adjust(value int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.Value += value
	if lm.Value > 100 {
		lm.Value = 100
	} else if lm.Value < 0 {
		lm.Value = 0
	}
}

func (lm *LoveMeter) GetValue() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.Value
}

// CatActions represents a collection of cat actions
type CatActions struct {
	actions []string
}

func NewCatActions() *CatActions {
	return &CatActions{
		actions: []string{"Meow!", "Purrs contentedly.", "Scratches the post.", "Hisses!", "Jumps around."},
	}
}

func (ca *CatActions) GetRandomAction() string {
	rand.Seed(time.Now().UnixNano())
	return ca.actions[rand.Intn(len(ca.actions))]
}

// RepeatedTimer executes a function at regular intervals
type RepeatedTimer struct {
	interval  time.Duration
	function  func()
	stopChan  chan bool
	isRunning bool
}

func NewRepeatedTimer(interval time.Duration, function func()) *RepeatedTimer {
	rt := &RepeatedTimer{
		interval:  interval,
		function:  function,
		stopChan:  make(chan bool),
		isRunning: false,
	}
	rt.Start()
	return rt
}

func (rt *RepeatedTimer) Start() {
	if rt.isRunning {
		return
	}
	rt.isRunning = true
	go func() {
		ticker := time.NewTicker(rt.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rt.function()
			case <-rt.stopChan:
				return
			}
		}
	}()
}

func (rt *RepeatedTimer) Stop() {
	if !rt.isRunning {
		return
	}
	rt.isRunning = false
	rt.stopChan <- true
	close(rt.stopChan)
	rt.stopChan = make(chan bool) // Reset for reuse
}

// CatBot represents the bot functionality
type CatBot struct {
	LoveMeter  *LoveMeter
	CatActions *CatActions
}

func NewCatBot() *CatBot {
	return &CatBot{
		LoveMeter:  NewLoveMeter(),
		CatActions: NewCatActions(),
	}
}

func (cb *CatBot) HandleCommand(command string, channel string) {
	switch command {
	case "!cat":
		cb.HandleCatCommand(channel)
	}
}

func (cb *CatBot) HandleCatCommand(channel string) {
	fmt.Printf("PRIVMSG %s :Meow!\n", channel)
}

func (cb *CatBot) Start(channel string) {
	NewRepeatedTimer(60*time.Second, func() {
		cb.HandleRandomAction(channel)
	})
}

func (cb *CatBot) HandleRandomAction(channel string) {
	action := cb.CatActions.GetRandomAction()
	fmt.Printf("PRIVMSG %s :%s\n", channel, action)
}

func main() {
	channel := "#example-channel"
	catBot := NewCatBot()

	// Start the repeated random action
	catBot.Start(channel)

	// Simulate user commands
	catBot.HandleCommand("!cat", channel)

	// Let the program run to observe the repeated random actions
	select {}
}
