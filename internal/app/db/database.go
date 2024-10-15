package db

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"tmff-discord-app/db"
	"tmff-discord-app/internal/app/config"
)

func SetupDatabase(conf *config.Config) (*sqlx.DB, error) {
	sqlDB, err := sql.Open("sqlite3", conf.DBFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open database")
	}
	dbx := sqlx.NewDb(sqlDB, "sqlite3")

	// Enable foreign key constraints
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, errors.Wrap(err, "could not enable foreign key constraints")
	}

	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "could not create sqlite3 driver")
	}

	s := bindata.Resource(db.AssetNames(), db.Asset)
	sourceDriver, err := bindata.WithInstance(s)
	if err != nil {
		return nil, errors.Wrap(err, "could not create bindata driver")
	}

	m, err := migrate.NewWithInstance("go-bindata", sourceDriver, "sqlite3", driver)
	if err != nil {
		return nil, errors.Wrap(err, "could not create migrate instance")
	}

	if upErr := m.Up(); upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		return nil, errors.Wrap(upErr, "could not migrate database")
	}
	return dbx, nil
}
