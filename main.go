package main

import (
	"catbot/internal/bot"
	"log"
)

func main() {
	if err := bot.StartBot(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}
