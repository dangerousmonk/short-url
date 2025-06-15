package database

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
)

// ApplyMigrations is used before application is up and running to apply required migrations to the database schema
func ApplyMigrations(cfg *config.Config) {
	logging.Log.Infof("DB DSN=%s", cfg.DatabaseDSN)
	migrations, err := migrate.New("file://migrations", cfg.DatabaseDSN)
	if err != nil {
		logging.Log.Fatalf("Failed to apply migrations: %v", err)
	}
	err = migrations.Up()

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logging.Log.Info("Migrations no change")
			return
		}
		log.Fatalf("Migrations failed: %v ", err)
		return
	}
	logging.Log.Info(" Migrations: success")
}
