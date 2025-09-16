// internal/services/commands/toplove.go
package commands

import (
	"context"
	"fmt"
	"strings"
)

// TopLove5Handler shows top 5 by LoveMeter from cat_player table.
// Register with: cmds.AddCommand("!toplove", cmds.(*commands.CommandControllerImpl).TopLove5Handler())
// internal/services/commands/toplove.go
func (c *CommandControllerImpl) TopLove5Handler() func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		if len(args) == 0 {
			return nil
		}
		msg := strings.TrimSpace(args[0])
		if !strings.HasPrefix(msg, "!toplove") {
			return nil
		}

		players, err := c.game.CatPlayerRepo.TopLoveMeter(ctx, c.game.Network, c.game.Channel, 5)
		if err != nil {
			// TEMP: surface the real error so we can fix it
			errMsg := fmt.Sprintf("toplove error: %v", err)
			fmt.Println("[toplove]", errMsg)
			c.game.IrcClient.Privmsg(c.game.Channel, errMsg)
			return err
		}

		if len(players) == 0 {
			c.game.IrcClient.Privmsg(c.game.Channel, "No love yet. Try `!pet purrito` first 😺")
			return nil
		}

		out := "💖😽 See who Purrito loves the most (Top 5): "
		for i, p := range players {
			if i > 0 {
				out += "  •  "
			}
			out += fmt.Sprintf("#%d %s (♥ %d)", i+1, p.Name, p.LoveMeter)
		}
		c.game.IrcClient.Privmsg(c.game.Channel, out)
		return nil
	}
}
