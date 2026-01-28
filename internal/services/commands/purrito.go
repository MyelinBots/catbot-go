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
			"",
			"✨ = How the game works = ✨",
			" * Pet, love, feed, catnip or laser with me to increase your Love Meter ❤️ (0–100%).",
			" * Reach 100% ❤️✨ to become Bonded —> this unlocks daily BondPoints 🌙✨",
			" * BondPoints are earned once per day while bonded (streaks give bonus points).",
			" * If you ignore me for a day, your bond may slowly fade 😿",
			" * Long bonding streaks unlock secret gifts and special titles 🎁",
			"",
			"🐾 = Commands you can use = 🐾",
			" * !pet purrito :::: Pet me, maybe I will purr... or scratch! 🐾",
			" * !love purrito :::: Show me some love... more love, more purrs 💗",
			" * !feed purrito :::: Feed me some tasty treats 🍣 🍗 🍤 🍉",
			" * !slap purrito :::: Tease me... but be careful 👋😼",
			" * !catnip purrito :::: Give me some catnip to boost my mood 🌿😸",
			" * !laser purrito :::: Find out when I was last seen chasing lasers 🔦⚡️",
			" * !status purrito :::: Check your love percentage, mood, and love bar ❤️😽",
			" * !toplove :::: See who I love the most 💖",
			"",
			"🌙 Tip: Come back every day to keep our bond strong and unlock rare rewards!",
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
