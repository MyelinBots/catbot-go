package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/services/catbot"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
)

type CommandController interface {
	HandleCommand(ctx context.Context, line *irc.Line) error
	AddCommand(command string, handler func(ctx context.Context, args ...string) error)
}

type CommandControllerImpl struct {
	game     catbot.CatBot
	commands map[string]func(ctx context.Context, args ...string) error
}

// Constructor
func NewCommandController(gameinstance catbot.CatBot) CommandController {
	return &CommandControllerImpl{
		game:     gameinstance,
		commands: make(map[string]func(ctx context.Context, args ...string) error),
	}
}

// HandleCommand parses an IRC line and dispatches to the correct handler
func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	if len(line.Args) < 2 {
		return nil
	}

	raw := line.Args[1] // e.g., "!pet purrito"
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0] // e.g., "!pet"
	args := parts[1:]
	fmt.Println("Handling command:", command)

	if handler, exists := c.commands[command]; exists {
		ctx = context_manager.SetNickContext(ctx, line.Nick)
		return handler(ctx, append([]string{command}, args...)...)
	}

	return nil
}

// AddCommand registers a command handler
func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler
}

// Exported wrapper function that delegates to catbot.HandleCatCommand
func WrapCatHandler(bot catbot.CatBot) func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		player := context_manager.GetNickFromContext(ctx)
		message := strings.Join(args, " ")
		bot.HandleCatCommand(ctx, player, message)
		return nil
	}
}
