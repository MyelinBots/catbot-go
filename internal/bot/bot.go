package bot

import (
	"catbot/config"
	"crypto/tls"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	irc "github.com/thoj/go-ircevent"
)

// CatActions handles the love meter and random responses for the cat
type CatActions struct {
	LoveMeter    int
	IsCatPresent bool
	Responses    []string
	sync.Mutex
}

// NewCatActions creates a new CatActions instance
func NewCatActions() *CatActions {
	return &CatActions{
		LoveMeter:    0,
		IsCatPresent: false,
		Responses: []string{
			"purr", "meow", "hiding", "hiss", "jump", "play", "scratch", "rub against", "surprise",
		},
	}
}

// AdjustLoveMeter adjusts the love meter for specific actions
func (ca *CatActions) AdjustLoveMeter(action string) string {
	ca.Lock()
	defer ca.Unlock()

	switch action {
	case "pet", "feed":
		ca.LoveMeter += 10
		if ca.LoveMeter > 100 {
			ca.LoveMeter = 100
		}
		return fmt.Sprintf("purrito purrs happily! Love Meter: %d", ca.LoveMeter)
	case "kick":
		ca.LoveMeter -= 10
		if ca.LoveMeter < 0 {
			ca.LoveMeter = 0
		}
		return fmt.Sprintf("purrito hisses angrily! Love Meter: %d", ca.LoveMeter)
	default:
		return "purrito doesn't respond."
	}
}

// GetRandomResponse generates a random cat response
func (ca *CatActions) GetRandomResponse() string {
	rand.Seed(time.Now().UnixNano())
	return ca.Responses[rand.Intn(len(ca.Responses))]
}

// StartBot starts the CatBot
func StartBot() error {
	cfg := config.LoadConfigOrPanic()

	fmt.Printf("Starting CatBot with config: %+v\n", cfg)

	// Create a new IRC client
	c := irc.IRC(cfg.IRCConfig.Nick, cfg.IRCConfig.Nick)
	c.UseTLS = cfg.IRCConfig.SSL
	c.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	serverAddress := fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)

	cat := NewCatActions()

	// Handle connected event
	c.AddCallback("001", func(e *irc.Event) {
		fmt.Printf("Connected to %s\n", cfg.IRCConfig.Host)
		for _, channel := range cfg.IRCConfig.Channels {
			c.Join(channel)
			fmt.Printf("Joined channel: %s\n", channel)
		}
	})

	// Handle PRIVMSG event
	c.AddCallback("PRIVMSG", func(e *irc.Event) {
		message := e.Message()
		user := e.Nick

		if strings.HasPrefix(message, "!") {
			command := strings.TrimPrefix(strings.ToLower(message), "!")

			if cat.IsCatPresent {
				switch command {
				case "pet":
					response := cat.AdjustLoveMeter("pet")
					c.Privmsg(e.Arguments[0], fmt.Sprintf("%s pets the cat. %s", user, response))
					c.Privmsg(e.Arguments[0], fmt.Sprintf("purrito response: %s", cat.GetRandomResponse()))
				case "feed":
					response := cat.AdjustLoveMeter("feed")
					c.Privmsg(e.Arguments[0], fmt.Sprintf("%s feeds the cat. %s", user, response))
					c.Privmsg(e.Arguments[0], fmt.Sprintf("purrito response: %s", cat.GetRandomResponse()))
				default:
					c.Privmsg(e.Arguments[0], fmt.Sprintf("%s: Unknown command. Try !pet or !feed.", user))
				}
			} else {
				c.Privmsg(e.Arguments[0], fmt.Sprintf("%s: purrito is not here right now.", user))
			}
		}
	})

	// Handle random cat appearances
	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(30)+10) * time.Second) // Random interval between 10-40 seconds
			cat.IsCatPresent = true
			for _, channel := range cfg.IRCConfig.Channels {
				c.Privmsg(channel, "purrito has appeared! You can !pet or !feed it!")
			}
			time.Sleep(15 * time.Second) // Cat stays for 15 seconds
			cat.IsCatPresent = false
			for _, channel := range cfg.IRCConfig.Channels {
				c.Privmsg(channel, "purrito has wandered off.")
			}
		}
	}()

	// Handle disconnection
	quit := make(chan bool)
	c.AddCallback("DISCONNECTED", func(e *irc.Event) {
		fmt.Println("Disconnected from IRC")
		quit <- true
	})

	// Connect to the server
	if err := c.Connect(serverAddress); err != nil { // Pass server address here
		fmt.Printf("Connection error: %s\n", err.Error())
		return err
	}

	<-quit
	return nil
}
