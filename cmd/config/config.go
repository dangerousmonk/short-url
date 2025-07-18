// Package config describes all the necessary constants and structures to run the application
package config

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	defaultServerAddr         = "localhost:8080"
	defaultBaseURL            = "http://localhost:8080"
	defaultLogLevel           = "INFO"
	defaultEnv                = "dev"
	defaultFilePath           = "./internal/repository/memory/storage.json"
	defaultMaxURLsBatchSize   = 5000
	defaultShutDownTimeout    = 15
	defaultJWTSecret          = "b6e2490a47c14cb7a1732aed3ba3f3c5"
	defaultCertPath           = "./cert.pem"
	defaultCertPrivateKeyPath = "./key.pem"
)

// Config represents a structure that contains all configurations options for the application.
type Config struct {
	ServerAddr         string `json:"server_address"`
	BaseURL            string `json:"base_url"`
	LogLevel           string
	Env                string
	StorageFilePath    string `json:"file_storage_path"`
	DatabaseDSN        string `json:"database_dsn"`
	JWTSecret          string
	CertPath           string
	CertPrivateKeyPath string
	JSONConfigFilePath string
	MaxURLsBatchSize   int
	ShutDownTimeout    int
	EnableHTTPS        bool `json:"enable_https"`
}

// InitConfig is used to initialize Config
func InitConfig() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		log.Printf("Unable to load envs from file %v", err)
	}
	// Read cmd
	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Address to run server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.StringVar(&cfg.StorageFilePath, "f", cfg.StorageFilePath, "Path to storage file")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.JSONConfigFilePath, "c", cfg.JSONConfigFilePath, "Path for json config")

	flag.Parse()

	// Read env
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

	enableHTTPS := os.Getenv("ENABLE_HTTPS")
	if enableHTTPS != "" {
		parsed, err := strconv.ParseBool(enableHTTPS)
		if err != nil {
			log.Fatal(err.Error())
		}
		cfg.EnableHTTPS = parsed
	}

	JSONConfig := os.Getenv("CONFIG")
	if JSONConfig != "" {
		cfg.JSONConfigFilePath = JSONConfig
	}

	// Read json config
	var jsonCfg Config
	if cfg.JSONConfigFilePath != "" {
		err := ParseJSONConfig(&jsonCfg, cfg.JSONConfigFilePath)
		if err != nil {
			log.Printf("Unable to load envs from json config %v", err)
		}
	}

	// Default envs
	if cfg.ServerAddr == "" && jsonCfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	} else if cfg.ServerAddr == "" && jsonCfg.ServerAddr != "" {
		cfg.ServerAddr = jsonCfg.ServerAddr
	}

	if cfg.BaseURL == "" && jsonCfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	} else if cfg.BaseURL == "" && jsonCfg.BaseURL != "" {
		cfg.BaseURL = jsonCfg.BaseURL
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
	if cfg.Env == "" {
		cfg.Env = defaultEnv
	}
	if cfg.StorageFilePath == "" && jsonCfg.StorageFilePath == "" {
		cfg.StorageFilePath = defaultFilePath
	} else if cfg.StorageFilePath == "" && jsonCfg.StorageFilePath != "" {
		cfg.StorageFilePath = jsonCfg.StorageFilePath
	}

	if cfg.CertPath == "" {
		cfg.CertPath = defaultCertPath
	}

	if cfg.CertPrivateKeyPath == "" {
		cfg.CertPrivateKeyPath = defaultCertPrivateKeyPath
	}

	if !cfg.EnableHTTPS {
		cfg.EnableHTTPS = jsonCfg.EnableHTTPS
	}

	if cfg.DatabaseDSN == "" && jsonCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonCfg.DatabaseDSN
	}

	cfg.MaxURLsBatchSize = defaultMaxURLsBatchSize
	cfg.JWTSecret = defaultJWTSecret
	cfg.ShutDownTimeout = defaultShutDownTimeout
	return cfg
}
