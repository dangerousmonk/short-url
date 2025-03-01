package config

import (
	"flag"
)

const (
	defaultServerAddr = "localhost:8080"
	defaultBaseURL    = "http://localhost:8080/"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

var Cfg *Config

func InitConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerAddr, "a", defaultServerAddr, "Address to run server")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base address for shortened URL")
	flag.Parse()

	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}

	Cfg = cfg
	return cfg
}
