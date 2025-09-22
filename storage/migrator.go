package storage

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrator struct {
	srcDriver source.Driver
	srcDB     database.Driver
}

//go:embed migrations/*.sql
var sqlFiles embed.FS

func NewMigrator(db *sql.DB) (*Migrator, error) {

	d, err := iofs.New(sqlFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("unable get migration files: %w", err)
	}
	postgresDrv, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to create db instance: %w", err)
	}
	return &Migrator{srcDriver: d, srcDB: postgresDrv}, nil
}

func (m *Migrator) ApplyMigrations(migrType string) error {

	migrateInstance, err := migrate.NewWithInstance("migration_embeded_sql_files", m.srcDriver, "psql_db", m.srcDB)
	if err != nil {
		return fmt.Errorf("unable to create migration: %v", err)
	}

	switch migrType {
	case "up":
		if err = migrateInstance.Up(); err != nil {
			return fmt.Errorf("unable to apply Up migrations %w", err)
		}
	case "down":
		if err = migrateInstance.Down(); err != nil {
			return fmt.Errorf("unable to apply Down migrations %w", err)
		}
	}

	m.srcDriver.Close()
	return nil
}
