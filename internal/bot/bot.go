package bot

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/MyelinBots/catbot-go/config"
	"github.com/MyelinBots/catbot-go/internal/db"
	"github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player"
	"github.com/MyelinBots/catbot-go/internal/healthcheck"
	"github.com/MyelinBots/catbot-go/internal/services/catbot"
	"github.com/MyelinBots/catbot-go/internal/services/commands"
	irc "github.com/fluffle/goirc/client"
)

type Identified struct {
	sync.Mutex
	identified bool
}

type GameInstances struct {
	sync.Mutex
	GameStarted      map[string]bool
	games            map[string]*catbot.CatBot
	commandInstances map[string]commands.CommandController
}

func StartBot() error {
	cfg := config.LoadConfigOrPanic()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	identified := &Identified{}

	fmt.Printf("Starting bot with config: %+v\n", cfg)
	healthcheck.StartHealthcheck(ctx, cfg.AppConfig)

	ircConfig := irc.NewConfig(cfg.IRCConfig.Nick)
	ircConfig.SSL = cfg.IRCConfig.SSL
	ircConfig.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	ircConfig.Server = fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)

	conn := irc.Client(ircConfig)

	gameInstances := &GameInstances{
		games:            make(map[string]*catbot.CatBot),
		commandInstances: make(map[string]commands.CommandController),
		GameStarted:      make(map[string]bool),
	}

	// helper to init a channel's game+commands in one place
	initChannel := func(channel string) error {
		// If you want a shared DB across channels, hoist this out of the func.
		database := db.NewDatabase(cfg.DBConfig)
		if database == nil || database.DB == nil {
			return fmt.Errorf("db init failed for channel %s", channel)
		}

		catPlayerRepository := cat_player.NewPlayerRepository(database)
		game := catbot.NewCatBot(conn, catPlayerRepository, cfg.IRCConfig.Network, channel)
		cmds := commands.NewCommandController(game)

		cmds.AddCommand("!pet", game.HandleCatCommand)
		cmds.AddCommand("!love", game.HandleCatCommand)
		cmds.AddCommand("!invite", commands.InviteHandler(conn))

		gameInstances.games[channel] = game
		gameInstances.commandInstances[channel] = cmds
		gameInstances.GameStarted[channel] = false
		return nil
	}

	// Preload configured channels
	for _, ch := range cfg.IRCConfig.Channels {
		gameInstances.Lock()
		err := initChannel(ch)
		gameInstances.Unlock()
		if err != nil {
			return err
		}
	}

	// Connected → join configured channels
	conn.HandleFunc(irc.CONNECTED, func(c *irc.Conn, _ *irc.Line) {
		fmt.Printf("Connected to %s\n", cfg.IRCConfig.Host)
		for _, ch := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", ch)
			c.Join(ch)
		}
	})

	// Also join on MOTD end / no MOTD (common IRC codes)
	conn.HandleFunc("422", func(c *irc.Conn, _ *irc.Line) {
		for _, ch := range cfg.IRCConfig.Channels {
			c.Join(ch)
		}
	})
	conn.HandleFunc("376", func(c *irc.Conn, _ *irc.Line) {
		for _, ch := range cfg.IRCConfig.Channels {
			c.Join(ch)
		}
	})

	// JOIN events: start the game loop for that channel (once)
	conn.HandleFunc(irc.JOIN, func(c *irc.Conn, line *irc.Line) {
		channel := line.Args[0]
		fmt.Printf("Joined %s\n", channel)

		// Identify with NickServ after we join somewhere
		handleNickserv(cfg.IRCConfig, identified, c)

		gameInstances.Lock()
		defer gameInstances.Unlock()

		// ensure the channel is initialized (covers invites or missing preload)
		if _, ok := gameInstances.games[channel]; !ok {
			if err := initChannel(channel); err != nil {
				fmt.Printf("Error init channel %s: %v\n", channel, err)
				return
			}
		}

		if !gameInstances.GameStarted[channel] {
			go gameInstances.games[channel].Start(ctx) // non-blocking
			gameInstances.GameStarted[channel] = true
		}
	})

	// INVITE: join and init the per-channel game/commands, then start
	conn.HandleFunc(irc.INVITE, func(c *irc.Conn, line *irc.Line) {
		channel := line.Args[1]
		fmt.Printf("Invited to %s\n", channel)
		c.Join(channel)

		gameInstances.Lock()
		defer gameInstances.Unlock()

		if err := initChannel(channel); err != nil {
			fmt.Printf("Error init invited channel %s: %v\n", channel, err)
			return
		}

		go gameInstances.games[channel].Start(ctx)
		gameInstances.GameStarted[channel] = true
	})

	// Command dispatcher
	conn.HandleFunc(irc.PRIVMSG, func(c *irc.Conn, line *irc.Line) {
		channel := line.Args[0]
		msg := line.Args[1]

		// explicit manual start
		if msg == "!start" {
			gameInstances.Lock()
			defer gameInstances.Unlock()

			game, ok := gameInstances.games[channel]
			if !ok {
				if err := initChannel(channel); err != nil {
					fmt.Printf("Error init channel %s: %v\n", channel, err)
					return
				}
				game = gameInstances.games[channel]
			}
			if gameInstances.GameStarted[channel] {
				fmt.Printf("Game already started for %s\n", channel)
				return
			}
			fmt.Printf("Starting gameInstance for %s\n", channel)
			go game.Start(ctx) // non-blocking
			gameInstances.GameStarted[channel] = true
			return
		}

		// Route to the channel's command controller
		gameInstances.Lock()
		cmds, ok := gameInstances.commandInstances[channel]
		gameInstances.Unlock()
		if !ok {
			// Lazy-init if needed
			gameInstances.Lock()
			if err := initChannel(channel); err != nil {
				gameInstances.Unlock()
				fmt.Printf("Error lazy init channel %s: %v\n", channel, err)
				return
			}
			cmds = gameInstances.commandInstances[channel]
			gameInstances.Unlock()
		}

		if err := cmds.HandleCommand(ctx, line); err != nil {
			fmt.Printf("Error handling command: %s\n", err.Error())
			return
		}
	})

	quit := make(chan bool, 1)
	conn.HandleFunc(irc.DISCONNECTED, func(_ *irc.Conn, _ *irc.Line) {
		quit <- true
	})

	if err := conn.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
		return err
	}

	<-quit
	return nil
}

func handleNickserv(cfg config.IRCConfig, identified *Identified, c *irc.Conn) {
	identified.Lock()
	defer identified.Unlock()
	if !identified.identified && cfg.NickservPassword != "" {
		command := fmt.Sprintf(cfg.NickservCommand, cfg.NickservPassword)
		c.Raw(command)
		identified.identified = true
	}
}
