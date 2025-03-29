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
	defaultFilePath   = "./internal/storage/storage.json"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	LogLevel        string
	Env             string
	StorageFilePath string
	DatabaseDSN     string
}

func InitConfig() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		log.Printf("Unable to load envs from file %v", err)
	}
	// Чтение флагов командной строки
	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Address to run server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.StringVar(&cfg.StorageFilePath, "f", cfg.StorageFilePath, "Path to storage file")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.Parse()

	// Чтение переменных окружения
	addr := os.Getenv("SERVER_ADDRESS")
	if addr != "" {
		cfg.ServerAddr = addr
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	envLogLevel := os.Getenv("LOG_LEVEL")
	if envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	storagePath := os.Getenv("FILE_STORAGE_PATH")
	if storagePath != "" {
		cfg.StorageFilePath = storagePath
	}

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN != "" {
		cfg.DatabaseDSN = dbDSN
	}

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
	if cfg.StorageFilePath == "" {
		cfg.StorageFilePath = defaultFilePath
	}
	return cfg
}
