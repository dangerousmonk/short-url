package config

import (
	"flag"
	"log"
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

func InitConfig() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		log.Printf("Unable to load envs from file %v", err)
	}

	addr := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	if addr != "" {
		cfg.ServerAddr = addr
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	if cfg.BaseURL != "" && cfg.ServerAddr != "" {
		return cfg
	}

	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Address to run server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.Parse()

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return cfg
}
