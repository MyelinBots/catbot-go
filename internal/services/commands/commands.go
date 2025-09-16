package commands

import (
	"context"
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
		// pass full message so handler can parse args
		return handler(ctx, message)
	}
	return nil
}

// AddCommand registers a command handler
func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler
}
