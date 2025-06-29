package commands

import (
	"context"
	"fmt"
	"github.com/MyelinBots/catbot-go/internal/services/catbot"
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
	"strings"
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
	// split by space and get the first element
	command := strings.Split(message, " ")[0]
	// args := line.Args[1:]
	fmt.Println("Handling command:", command)

	if handler, exists := c.commands[command]; exists {
		fmt.Println("Handling command:", command)
		ctx = context_manager.SetNickContext(ctx, line.Nick)
		return handler(ctx, line.Args[1:]...)
	} else {
		return nil
	}
}

// AddCommand registers a command handler
func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler
}
