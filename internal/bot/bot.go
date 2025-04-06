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
	"github.com/MyelinBots/catbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
)

type GameStarted struct {
	sync.Mutex
	started bool
}

type Identified struct {
	sync.Mutex
	identified bool
}

func StartBot() error {
	cfg := config.LoadConfigOrPanic()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := &GameStarted{}
	identified := &Identified{}

	fmt.Printf("Starting bot with config: %+v\n", cfg)
	healthcheck.StartHealthcheck(ctx, cfg.AppConfig)

	ircConfig := irc.NewConfig(cfg.IRCConfig.Nick)
	ircConfig.SSL = cfg.IRCConfig.SSL
	ircConfig.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	ircConfig.Server = fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)

	conn := irc.Client(ircConfig)

	cat := catbot.NewCatBot(conn, cfg.IRCConfig.Channels[0])
	controller := commands.NewCommandController(cat)

	controller.AddCommand("!pet", commands.WrapCatHandler(cat))
	controller.AddCommand("!love", commands.WrapCatHandler(cat))
	controller.AddCommand("!invite", commands.InviteHandler(conn))

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
		started.Lock()
		defer started.Unlock()
		if line.Args[0] == cfg.IRCConfig.Channels[0] && !started.started {
			started.started = true
		}
		handleNickserv(cfg.IRCConfig, identified, conn)
	})

	conn.HandleFunc(irc.INVITE, func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("Invited to %s\n", line.Args[1])
		conn.Join(line.Args[1])
	})

	conn.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		if line == nil || len(line.Args) < 2 {
			return
		}

		ctx := context_manager.SetNickContext(context.Background(), line.Nick)
		ctx = context_manager.SetLineContext(ctx, line)

		err := controller.HandleCommand(ctx, line)
		if err != nil {
			fmt.Printf("Error handling command: %s\n", err.Error())
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
