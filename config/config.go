package config

import (
	"strings"

	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig AppConfig `env:"APPCONFIG"`
	IRCConfig IRCConfig `env:"IRCCONFIG"`
	DBConfig  DBConfig  `env:"DBCONFIG"`
}

type AppConfig struct {
	APPName string `default:"purrito"`
	Version string `default:"x.x.x" env:"VERSION"`
	Port    int    `default:"8080" env:"APP_PORT"`
}

type IRCConfig struct {
	Host             string `env:"HOST"`
	Port             int    `env:"PORT"`
	SSL              bool   `env:"SSL"`
	Nick             string `env:"NICK"`
	ChannelsString   string `env:"CHANNELS"`
	Channels         []string
	Network          string `env:"NETWORK"`
	NickservCommand  string `env:"NICKSERV_COMMAND" default:"PRIVMSG NickServ IDENTIFY %s"`
	NickservPassword string `env:"NICKSERV_PASSWORD" default:""`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"purrito" env:"DBNAME"`
	User     string `default:"postgres" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"5432" env:"DBPORT"`
	SSLMode  string `default:"disable" env:"DBSSL"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	config.IRCConfig.Channels = strings.Split(config.IRCConfig.ChannelsString, ",")

	return config
}
