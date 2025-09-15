package main

import (
	"log"

	"github.com/MyelinBots/catbot-go/config"
	"github.com/MyelinBots/catbot-go/internal/bot"
	"github.com/MyelinBots/catbot-go/internal/db"
)

func main() {
	cfg := config.LoadConfigOrPanic()

	database := db.NewDatabase(cfg.DBConfig)
	if database == nil || database.DB == nil {
		log.Fatalf("failed to connect to database")
	}

	// Just start the bot, no args
	if err := bot.StartBot(); err != nil {
		log.Fatalf("error starting bot: %v", err)
	}
}
