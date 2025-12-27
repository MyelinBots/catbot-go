package commands

import (
	"context"
	"math/rand"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/catbot"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
)

type CommandController interface {
	HandleCommand(ctx context.Context, line *irc.Line) error
	AddCommand(command string, handler func(ctx context.Context, message string) error)
}

type CommandControllerImpl struct {
	game     *catbot.CatBot
	commands map[string]func(ctx context.Context, message string) error
}

func NewCommandController(gameinstance *catbot.CatBot) CommandController {
	return &CommandControllerImpl{
		game:     gameinstance,
		commands: make(map[string]func(ctx context.Context, message string) error),
	}
}

// HandleCommand parses an IRC line and dispatches to the correct handler
func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	if len(line.Args) < 2 {
		return nil
	}

	message := line.Args[1]
	command := strings.Fields(message)
	if len(command) == 0 {
		return nil
	}

	cmd := command[0]
	if handler, exists := c.commands[cmd]; exists {
		ctx = context_manager.SetNickContext(ctx, line.Nick)
		return handler(ctx, message)
	}
	return nil
}

func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, message string) error) {
	c.commands[command] = handler
}

// PurritoLaserHandler: handles ONLY "!laser purrito"
func (c *CommandControllerImpl) PurritoLaserHandler() func(ctx context.Context, message string) error {
	return func(ctx context.Context, message string) error {
		nick := context_manager.GetNickContext(ctx)

		parts := strings.Fields(strings.TrimSpace(message))
		// Expect: !laser purrito
		if len(parts) < 2 || !strings.EqualFold(parts[0], "!laser") || !strings.EqualFold(parts[1], "purrito") {
			return nil
		}

		// Require Purrito to be present (time window)
		if !c.game.IsPresent() {
			c.game.IrcClient.Privmsg(c.game.Channel, "🐾 Purrito is not here right now. Wait until he shows up!")
			return nil
		}

		laserMoves := []string{
			"🔦⚡️ The laser flickers! Purrito darts after it, paws flying everywhere!",
			"🔦⚡️ Purrito spots the laser and wiggles — then pounces!",
			"🔦⚡️ Purrito chases the laser dot in circles... dizzy but happy!",
			"🔦⚡️ Purrito dives at the laser, misses, then looks proud anyway.",
			"🔦⚡️ The red dot dances — Purrito bats at it with lightning speed!",
			"🔦⚡️ Purrito takes a break, watching the laser with intense focus.",
		}
		c.game.IrcClient.Privmsg(c.game.Channel, laserMoves[rand.Intn(len(laserMoves))])

		// Optional: small love boost (safe assert)
		if ca, ok := c.game.CatActions.(*cat_actions.CatActions); ok {
			ca.LoveMeter.Increase(nick, 1)
		}

		// Optional: if you still want laser to count as "interaction" for the leave message,
		// you have two options:
		// 1) Put "laser" inside needsPurritoPresent in CatBot.HandleCatCommand (recommended)
		// 2) Or add a public method in catbot: MarkInteracted() and call it here.
		return nil
	}
}
