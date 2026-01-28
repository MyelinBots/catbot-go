package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/services/cat_actions"
	"github.com/MyelinBots/catbot-go/internal/services/catbot"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
)

// --------------------------------------------------
// Interfaces
// --------------------------------------------------

type CommandController interface {
	HandleCommand(ctx context.Context, line *irc.Line) error
	AddCommand(command string, handler func(ctx context.Context, message string) error)
}

// --------------------------------------------------
// Controller
// --------------------------------------------------

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

// --------------------------------------------------
// Core dispatcher
// --------------------------------------------------

func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	if len(line.Args) < 2 {
		return nil
	}

	message := line.Args[1]
	fields := strings.Fields(message)
	if len(fields) == 0 {
		return nil
	}

	cmd := fields[0]
	handler, exists := c.commands[cmd]
	if !exists {
		return nil
	}

	ctx = context_manager.SetNickContext(ctx, line.Nick)
	return handler(ctx, message)
}

func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, message string) error) {
	c.commands[command] = handler
}

// --------------------------------------------------
// Shared helper: Bonded streak + BondPoints + Total
// --------------------------------------------------

func (c *CommandControllerImpl) appendBondProgress(ctx context.Context, nick string, msg string) string {
	// ✅ Do NOT append anything to the catnip cooldown message
	if strings.Contains(strings.ToLower(msg), "already used catnip today") {
		return msg
	}

	ca, ok := c.game.CatActions.(*cat_actions.CatActions)
	if !ok || ca.LoveMeter == nil {
		return msg
	}

	if ca.LoveMeter.Get(nick) != 100 {
		return msg
	}

	pts, streak, err := ca.LoveMeter.RecordInteraction(ctx, nick)
	if err != nil {
		return msg
	}

	p, err := ca.CatPlayerRepo.GetPlayerByName(ctx, nick, ca.Network, ca.Channel)
	total := 0
	if err == nil && p != nil {
		total = p.BondPoints
	}

	if pts > 0 {
		return msg + fmt.Sprintf(
			" :: Bonded streak: %d day(s) :: +%d BondPoints (Total: %d)",
			streak, pts, total,
		)
	}

	return msg + fmt.Sprintf(" :: BondPoints already earned today (Total: %d)", total)
}

// --------------------------------------------------
// Handlers
// --------------------------------------------------

// PurritoLaserHandler: handles ONLY "!laser purrito"
// CatActions.ExecuteAction handles presence gating, love changes, and message formatting.
func (c *CommandControllerImpl) PurritoLaserHandler() func(ctx context.Context, message string) error {
	return func(ctx context.Context, message string) error {
		nick := context_manager.GetNickContext(ctx)

		parts := strings.Fields(strings.TrimSpace(message))
		if len(parts) < 2 || !strings.EqualFold(parts[0], "!laser") {
			return nil
		}
		if !strings.EqualFold(parts[1], "purrito") {
			return nil
		}

		// CatActions handles everything: presence check, love changes, message formatting
		out := c.game.CatActions.ExecuteAction("laser", nick, "purrito")
		out = c.appendBondProgress(ctx, nick, out)
		c.game.IrcClient.Privmsg(c.game.Channel, out)
		return nil
	}
}
