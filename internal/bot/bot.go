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

	// ---- IRC config (with PASS) ----
	ircConfig := irc.NewConfig(cfg.IRCConfig.Nick)
	ircConfig.SSL = cfg.IRCConfig.SSL
	ircConfig.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	ircConfig.Server = fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)
	ircConfig.Me.Ident = cfg.IRCConfig.User
	ircConfig.Pass = cfg.IRCConfig.Password // send PASS to bouncer

	conn := irc.Client(ircConfig)

	// ---- DB: open ONCE and migrate ----
	database := db.NewDatabase(cfg.DBConfig)
	if database == nil || database.DB == nil {
		return fmt.Errorf("db init failed")
	}
	if err := database.DB.AutoMigrate(&cat_player.CatPlayer{}); err != nil {
		return fmt.Errorf("migrate cat_player failed: %w", err)
	}

	gameInstances := &GameInstances{
		games:            make(map[string]*catbot.CatBot),
		commandInstances: make(map[string]commands.CommandController),
		GameStarted:      make(map[string]bool),
	}

	// helper to init a channel's game+commands in one place (REUSE 'database')
	initChannel := func(channel string) error {
		catPlayerRepository := cat_player.NewPlayerRepository(database)
		game := catbot.NewCatBot(conn, catPlayerRepository, cfg.IRCConfig.Network, channel)
		cmds := commands.NewCommandController(game)

		cmds.AddCommand("!pet", game.HandleCatCommand)
		cmds.AddCommand("!love", game.HandleCatCommand)
		cmds.AddCommand("!invite", commands.InviteHandler(conn))
		cmds.AddCommand("!toplove", cmds.(*commands.CommandControllerImpl).TopLove5Handler())
		cmds.AddCommand("!purrito", cmds.(*commands.CommandControllerImpl).PurritoHandler())

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

	// Also join on MOTD end / no MOTD
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

		handleNickserv(cfg.IRCConfig, identified, c)

		gameInstances.Lock()
		defer gameInstances.Unlock()

		if _, ok := gameInstances.games[channel]; !ok {
			if err := initChannel(channel); err != nil {
				fmt.Printf("Error init channel %s: %v\n", channel, err)
				return
			}
		}
		if !gameInstances.GameStarted[channel] {
			gameInstances.GameStarted[channel] = true
			go gameInstances.games[channel].Start(ctx)
		}
	})

	// INVITE handler
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
			go game.Start(ctx)
			gameInstances.GameStarted[channel] = true
			return
		}

		gameInstances.Lock()
		cmds, ok := gameInstances.commandInstances[channel]
		gameInstances.Unlock()
		if !ok {
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
	conn.HandleFunc(irc.DISCONNECTED, func(_ *irc.Conn, _ *irc.Line) { quit <- true })

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
		// Must include ':' before IDENTIFY param
		command := fmt.Sprintf("PRIVMSG NickServ :IDENTIFY %s", cfg.NickservPassword)
		c.Raw(command)
		identified.identified = true
	}
}
