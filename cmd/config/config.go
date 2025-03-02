package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
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
	godotenv.Load()

	flag.StringVar(&Cfg.ServerAddr, "a", defaultServerAddr, "Address to run server")
	flag.StringVar(&Cfg.BaseURL, "b", defaultBaseURL, "Base address for shortened URL")
	flag.Parse()

	addr := os.Getenv("SERVER_ADDRESS")
	basURL := os.Getenv("BASE_URL")
	if addr != "" {
		Cfg.ServerAddr = addr
	}
	if basURL != "" {
		Cfg.BaseURL = basURL
	}

	return Cfg
}
