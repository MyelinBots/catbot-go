package bot

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/MyelinBots/catbot-go/config"
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

	for _, channel := range cfg.IRCConfig.Channels {
		gameInstances.Lock()
		gameInstance := catbot.NewCatBot(conn, channel)
		commandInstance := commands.NewCommandController(gameInstance)

		commandInstance.AddCommand("!pet", gameInstance.HandleCatCommand)
		commandInstance.AddCommand("!love", gameInstance.HandleCatCommand)
		commandInstance.AddCommand("!invite", commands.InviteHandler(conn))

		gameInstances.games[channel] = gameInstance
		gameInstances.commandInstances[channel] = commandInstance
		gameInstances.GameStarted[channel] = false
		gameInstances.Unlock()
	}

	conn.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("Connected to %s\n", cfg.IRCConfig.Host)
		for _, channel := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", channel)
			conn.Join(channel)
		}
	})

	conn.HandleFunc("422", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}
	})

	conn.HandleFunc("376", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}
	})

	conn.HandleFunc(irc.JOIN, func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("Joined %s\n", line.Args[0])
		gameInstances.Lock()

		if gameInstance, ok := gameInstances.games[line.Args[0]]; ok {
			gameInstance := gameInstance

			if !gameInstances.GameStarted[line.Args[0]] {
				go gameInstance.Start(ctx)
				gameInstances.GameStarted[line.Args[0]] = true
			}

		}
		gameInstances.Unlock()
		//// if channel is first channel in config
		//if line.Args[0] == cfg.IRCConfig.Channels[0] && !started.started {
		//	go gameInstance.Start(ctx)
		//	started.started = true
		//}

		handleNickserv(cfg.IRCConfig, identified, conn)
		return

	})

	conn.HandleFunc(irc.INVITE, func(conn *irc.Conn, line *irc.Line) {
		channel := line.Args[1]
		fmt.Printf("Invited to %s\n", line.Args[1])
		conn.Join(line.Args[1])
		gameInstances.Lock()
		gameInstance := catbot.NewCatBot(conn, channel)
		commandInstance := commands.NewCommandController(gameInstance)

		commandInstance.AddCommand("!pet", gameInstance.HandleCatCommand)
		commandInstance.AddCommand("!love", gameInstance.HandleCatCommand)
		commandInstance.AddCommand("!invite", commands.InviteHandler(conn))

		gameInstances.games[channel] = gameInstance
		gameInstances.commandInstances[channel] = commandInstance
		gameInstances.GameStarted[channel] = false

		// start game instance
		go gameInstance.Start(ctx)
		gameInstances.GameStarted[channel] = true
		gameInstances.Unlock()

	})

	conn.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		// if message is !shoot
		// if message is !start
		if line.Args[1] == "!start" {

			gameInstances.Lock()
			if gameInstance, ok := gameInstances.games[line.Args[0]]; ok {

				gameInstances.Unlock()

				if gameInstances.GameStarted[line.Args[0]] {
					fmt.Printf("Game already started for %s\n", line.Args[0])
					return
				}

				fmt.Printf("Starting gameInstance for %s\n", line.Args[0])
				gameInstance.Start(ctx)
				gameInstances.GameStarted[line.Args[0]] = true
				return
			}

		}
		gameInstances.Lock()
		commandInstance := gameInstances.commandInstances[line.Args[0]]
		gameInstances.Unlock()
		err := commandInstance.HandleCommand(ctx, line)
		if err != nil {
			fmt.Printf("Error handling command: %s\n", err.Error())
			return
		}

	})

	quit := make(chan bool)
	conn.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
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
