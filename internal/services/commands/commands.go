package commands

import (
	"context"
	"fmt"
	"math/rand"
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
// ✅ Route to CatActions so presence/catnip rules are consistent everywhere.
// ✅ IMPORTANT: do NOT re-roll accept/reject here (CatActions already decided).
func (c *CommandControllerImpl) PurritoLaserHandler() func(ctx context.Context, message string) error {
	acceptMoves := []string{
		"🔦⚡️ The laser flickers! Purrito darts after it, paws flying everywhere!",
		"🔦⚡️ Purrito spots the laser and wiggles ... then pounces!",
		"🔦⚡️ Purrito chases the laser dot in circles... dizzy but happy!",
		"🔦⚡️ Purrito dives at the laser, misses, then looks proud anyway.",
		"🔦⚡️ The red dot dances ... Purrito bats at it with lightning speed!",
		"🔦⚡️ Purrito takes a break, watching the laser with intense focus.",
	}

	rejectMoves := []string{
		"🔦😾 Purrito narrows his eyes... not impressed by the laser right now.",
		"🔦🙄 Purrito ignores the dot and grooms his paw instead.",
		"🔦😿 Purrito flops down ... too tired to chase today.",
		"🔦😼 Purrito watches... then turns away like it’s beneath him.",
		"🔦😾 Purrito swishes his tail in annoyance and refuses to play.",
	}

	return func(ctx context.Context, message string) error {
		nick := context_manager.GetNickContext(ctx)

		parts := strings.Fields(strings.TrimSpace(message))
		if len(parts) < 2 || !strings.EqualFold(parts[0], "!laser") {
			return nil
		}
		if !strings.EqualFold(parts[1], "purrito") {
			return nil
		}

		ca, ok := c.game.CatActions.(*cat_actions.CatActions)
		if !ok || ca.LoveMeter == nil {
			out := c.game.CatActions.ExecuteAction("laser", nick, "purrito")
			out = c.appendBondProgress(ctx, nick, out)
			c.game.IrcClient.Privmsg(c.game.Channel, out)
			return nil
		}

		loveBefore := ca.LoveMeter.Get(nick)

		// Single source of truth: presence/ownership handled by CatActions.
		out := c.game.CatActions.ExecuteAction("laser", nick, "purrito")

		// If blocked ("not here"), forward as-is (no decoration)
		if strings.Contains(out, "🐾 Purrito is not here right now") {
			out = c.appendBondProgress(ctx, nick, out)
			c.game.IrcClient.Privmsg(c.game.Channel, out)
			return nil
		}

		// Decorate based on real love change from CatActions
		loveAfter := ca.LoveMeter.Get(nick)
		mood := ca.LoveMeter.GetMood(nick)
		bar := ca.LoveMeter.GetLoveBar(nick)

		if loveAfter > loveBefore {
			msg := acceptMoves[rand.Intn(len(acceptMoves))]
			out = fmt.Sprintf("%s Your love meter is now %d%% and purrito is now %s %s", msg, loveAfter, mood, bar)
		} else {
			msg := rejectMoves[rand.Intn(len(rejectMoves))]
			out = fmt.Sprintf("%s Your love meter is now %d%% and purrito is now %s %s", msg, loveAfter, mood, bar)
		}

		out = c.appendBondProgress(ctx, nick, out)
		c.game.IrcClient.Privmsg(c.game.Channel, out)
		return nil
	}
}
