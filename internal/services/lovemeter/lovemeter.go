package lovemeter

import (
	"context"
	"fmt"
	"strings"

	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
)

type LoveMeter interface {
	Increase(player string, amount int)
	Decrease(player string, amount int)
	Get(player string) int
	GetLoveBar(player string) string
}

type LoveMeterImpl struct {
	Values        map[string]int
	CatPlayerRepo cat_player.CatPlayerRepository
	Network       string
	Channel       string
}

func NewLoveMeter(catPlayerRepo cat_player.CatPlayerRepository, network string, channel string) LoveMeter {
	return &LoveMeterImpl{
		Values:        make(map[string]int),
		CatPlayerRepo: catPlayerRepo,
		Network:       network,
		Channel:       channel,
	}
}

func (lm *LoveMeterImpl) Increase(player string, amount int) {
	lm.Values[player] += amount
	if lm.Values[player] > 100 {
		lm.Values[player] = 100
	}
	// Update the player's love meter in the database
	if err := lm.CatPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      player,
		LoveMeter: lm.Values[player],
		Network:   lm.Network,
		Channel:   lm.Channel,
	}); err != nil {
		fmt.Printf("Error updating love meter for player %s: %v\n", player, err)
	}
}

func (lm *LoveMeterImpl) Decrease(player string, amount int) {
	lm.Values[player] -= amount
	if lm.Values[player] < 0 {
		lm.Values[player] = 0
	}
	// Update the player's love meter in the database
	if err := lm.CatPlayerRepo.UpsertPlayer(context.Background(), &cat_player.CatPlayer{
		Name:      player,
		LoveMeter: lm.Values[player],
		Network:   lm.Network,
		Channel:   lm.Channel,
	}); err != nil {
		fmt.Printf("Error updating love meter for player %s: %v\n", player, err)
	}
}

func (lm *LoveMeterImpl) Get(player string) int {
	// get from db
	fPlayer, err := lm.CatPlayerRepo.GetPlayerByName(player, lm.Network, lm.Channel)
	if err != nil {
		fmt.Printf("Error getting player %s: %v\n", player, err)
		return 0
	}
	if fPlayer == nil {
		// Player not found, initialize with 0 love
		lm.Values[fPlayer.Name] = 0
		return 0
	} else {
		// Player found, return their love meter value
		lm.Values[fPlayer.Name] = fPlayer.LoveMeter
		return fPlayer.LoveMeter
	}
}

func (lm *LoveMeterImpl) GetLoveBar(player string) string {
	love := lm.Get(player)
	return fmt.Sprintf("%s", lm.generateLoveBar(love))
}

func (lm *LoveMeterImpl) generateLoveBar(percent int) string {
	// Generate a love bar based on the percentage
	// This is a placeholder implementation
	// show percentage as a bar of hearts with a total of 10 hearts being 100%
	hearts := percent / 10
	return fmt.Sprintf("[%s%s]", strings.Repeat("❤️", hearts), strings.Repeat("░", 10-hearts))
}

func (lm *LoveMeterImpl) GetMood(player string) string {
	love := lm.Get(player)
	switch {
	case love == 0:
		return "😾 hostile"
	case love < 20:
		return "😿 sad"
	case love < 50:
		return "😐 cautious"
	case love < 80:
		return "😺 friendly"
	default:
		return "😻 loves you"
	}
}
