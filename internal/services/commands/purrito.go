package commands

import (
	"context"

	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
)

func (c *CommandControllerImpl) PurritoHandler() func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		nick := context_manager.GetNickContext(ctx)

		lines := []string{
			"🐱 Hi " + nick + "! I am \x0303Purrito\x0F — your friendly IRC cat on the \x0311DarkWorld Network\x0F",
			"",
			"\x0310✨ = How the game works = ✨\x0F",
			"\x0309 * \x0FPet, love, feed, catnip or laser with me to increase your \x0313Love Meter\x0F ❤️ \x0311(0–100%)\x0F",
			"\x0309 * \x0FReach \x0303100%\x0F ❤️ to become \x0313Bonded\x0F —> this unlocks \x0310daily BondPoints\x0F ⭐",
			"\x0309 * \x0FBondPoints are earned \x0311once per day\x0F while bonded \x0307(streaks give bonus points)\x0F",
			"\x0309 * \x0FIf you ignore me for a day, your bond may slowly fade... \x0304</3\x0F",
			"\x0309 * \x0FLong bonding streaks unlock \x0313secret gifts\x0F and \x0310special titles\x0F 🎁",
			"",
			"\x0310🐾 = Commands you can use = 🐾\x0F",
			"\x0311 * \x0F!pet purrito \x0307::::\x0F Pet me, maybe I will purr... or scratch! 🐾",
			"\x0311 * \x0F!love purrito \x0307::::\x0F Show me some love... more love, more purrs 💗",
			"\x0311 * \x0F!feed purrito \x0307::::\x0F Feed me some tasty treats 🍣 🍗 🍤 🍉",
			"\x0311 * \x0F!slap purrito \x0307::::\x0F Tease me... but be careful 👋😼",
			"\x0311 * \x0F!catnip purrito \x0307::::\x0F Give me some catnip to boost my mood 🌿😸",
			"\x0311 * \x0F!laser purrito \x0307::::\x0F Find out when I was last seen chasing lasers 🔦⚡️",
			"\x0311 * \x0F!status purrito \x0307::::\x0F Check your love, mood, bond & gifts ❤️😽",
			"\x0311 * \x0F!toplove \x0307::::\x0F See who I love the most 💖",
			"",
			"\x0313= Tip =\x0F Come back \x0311every day\x0F to keep our bond strong and unlock \x0303rare rewards\x0F ✨",
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
