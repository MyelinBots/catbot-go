package main

import (
	"log"

	"github.com/MyelinBots/catbot-go/internal/bot"
)

func main() {
	if err := bot.StartBot(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}
