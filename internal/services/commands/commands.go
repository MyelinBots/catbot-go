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
	AddCommand(command string, handler func(ctx context.Context, args ...string) error)
}

type CommandControllerImpl struct {
	game     *catbot.CatBot
	commands map[string]func(ctx context.Context, args ...string) error
}

// Constructor
func NewCommandController(gameinstance *catbot.CatBot) CommandController {
	return &CommandControllerImpl{
		game:     gameinstance,
		commands: make(map[string]func(ctx context.Context, args ...string) error),
	}
}

// HandleCommand parses an IRC line and dispatches to the correct handler
func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	message := line.Args[1]
	command := strings.Split(message, " ")[0]
	if handler, exists := c.commands[command]; exists {
		ctx = context_manager.SetNickContext(ctx, line.Nick)
		// pass the full message so handlers can parse args if needed
		return handler(ctx, message)
	}
	return nil
}

// AddCommand registers a command handler
func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler
}

// PurritoLaserHandler shows when Purrito was last seen and when he'll appear again.
func (c *CommandControllerImpl) PurritoLaserHandler() func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		nick := context_manager.GetNickContext(ctx)
		if len(args) == 0 {
			return nil
		}
		msg := strings.TrimSpace(args[0])
		parts := strings.Fields(msg)
		if len(parts) < 2 || !strings.HasPrefix(parts[0], "!laser") || !strings.EqualFold(parts[1], "purrito") {
			return nil
		}

		// Require Purrito to be present
		if !c.game.IsPresent() {
			c.game.IrcClient.Privmsg(c.game.Channel, "🐾 Purrito is not here right now. Wait until he shows up!")
			return nil
		}

		// Count as an interaction for the current appearance window
		c.game.MarkInteracted()

		// Random playful laser reactions
		laserMoves := []string{
			"🔦⚡️ The laser flickers! Purrito darts after it, paws flying everywhere!",
			"🔦⚡️ Purrito spots the laser and wiggles — then pounces!",
			"🔦⚡️ Purrito chases the laser dot in circles... dizzy but happy!",
			"🔦⚡️ Purrito dives at the laser, misses, then looks proud anyway.",
			"🔦⚡️ The red dot dances — Purrito bats at it with lightning speed!",
			"🔦⚡️ Purrito takes a break, watching the laser with intense focus.",
		}
		c.game.IrcClient.Privmsg(c.game.Channel, laserMoves[rand.Intn(len(laserMoves))])

		// Small love boost for playful interaction (guard the type assert)
		if ca, ok := c.game.CatActions.(*cat_actions.CatActions); ok {
			ca.LoveMeter.Increase(nick, 1)
		}
		return nil
	}
}
