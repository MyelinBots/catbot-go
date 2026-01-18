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
			" * !pet purrito :::: Pet me, maybe I will purr... or scratch! 🐾",
			" * !love purrito :::: Show me some love... more love, more purrs 💗",
			" * !feed purrito :::: Feed me some tasty treats 🍣 🍗 🍤 🍉",
			" * !slap purrito :::: Tease me... but be careful 👋😼",
			" * !catnip purrito :::: Give me some catnip to boost my mood 🌿😸",
			" * !laser purrito :::: Find out when I was last seen chasing lasers 🔦⚡️",
			" * !status purrito :::: Check your love percentage, mood, and love bar ❤️😽",
			" * !purrito :::: Learn who I am and how the game works 🐱",
			" * !toplove :::: See who I love the most 💖",
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
