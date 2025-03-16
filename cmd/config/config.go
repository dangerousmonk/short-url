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
	defaultLogLevel   = "INFO"
	defaultEnv        = "dev"
)

type Config struct {
	ServerAddr string
	BaseURL    string
	LogLevel   string
	Env        string
}

func InitConfig() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		log.Printf("Unable to load envs from file %v", err)
	}

	// Чтение переменных окружения
	addr := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	envLogLevel := os.Getenv("LOG_LEVEL")
	if addr != "" {
		cfg.ServerAddr = addr
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	// Чтение флагов командной строки
	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Address to run server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.Parse()

	// Инициализация переменных по умолчанию
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
	if cfg.Env == "" {
		cfg.Env = defaultEnv
	}
	return cfg
}
