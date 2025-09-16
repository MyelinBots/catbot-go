package commands

import (
	"context"

	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

func (c *CommandControllerImpl) PurritoHandler() func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		nick := context_manager.GetNickContext(ctx)

		lines := []string{
			"🐱 Hi " + nick + "! I am Purrito — your friendly IRC cat on the DarkWorld Network.",
			"You can interact with me using commands:",
			" • !pet purrito — Pet me, maybe I will purr... or scratch! 🐾",
			" • !toplove — See who I love the most 💖",
			" • !purrito — Show this help/introduction",
		}

		for _, l := range lines {
			// keep each message reasonably short to avoid server truncation
			if len(l) > 400 {
				l = l[:400]
			}
			c.game.IrcClient.Privmsg(c.game.Channel, l)
		}
		return nil
	}
}
