package config

import (
	"flag"
)

const (
	defaultServerAddr = "localhost:8080"
	defaultBaseURL    = "http://localhost:8080"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

var Cfg *Config = &Config{}

func InitConfig() *Config {
	flag.StringVar(&Cfg.ServerAddr, "a", defaultServerAddr, "Address to run server")
	flag.StringVar(&Cfg.BaseURL, "b", defaultBaseURL, "Base address for shortened URL")
	flag.Parse()

	if Cfg.BaseURL == "" {
		Cfg.BaseURL = defaultBaseURL
	}

	if Cfg.ServerAddr == "" {
		Cfg.ServerAddr = defaultServerAddr
	}

	return Cfg
}
