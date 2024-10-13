package repository_test

import (
	"database/sql"
	"log"
	"testing"
	"tmff-discord-app/pkg/database"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func newMigratedSQLiteDB(t *testing.T) *sqlx.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	dbx := sqlx.NewDb(db, "sqlite3")

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign key constraints:", err)
	}

	var foreignKeysEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeysEnabled)
	if err != nil {
		log.Fatal("Failed to verify foreign key constraints:", err)
	}
	if foreignKeysEnabled == 0 {
		log.Fatal("Foreign key constraints are not enabled")
	}

	err = database.Migrate(db, "./../../../db/migrations")
	require.NoError(t, err)
	return dbx
}
