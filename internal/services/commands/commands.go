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

func NewCommandController(gameinstance catbot.CatBot) CommandController {
	return &CommandControllerImpl{
		game:     gameinstance,
		commands: make(map[string]func(ctx context.Context, args ...string) error),
	}
}

func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	if len(line.Args) < 2 {
		return nil
	}

	raw := line.Args[1] // e.g., "!pet purrito"
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]
	fmt.Println("Handling command:", command)

	if handler, exists := c.commands[command]; exists {
		ctx = context_manager.SetNickContext(ctx, line.Nick)
		return handler(ctx, args...)
	}

	return nil
}

func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler
}
