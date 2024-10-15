package repository_test

import (
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
	"tmff-discord-app/internal/app/config"
	"tmff-discord-app/internal/app/db"
)

func newMigratedSQLiteDB(t *testing.T) *sqlx.DB {
	conf := &config.Config{
		DBFile: ":memory:",
	}
	dbx, err := db.SetupDatabase(conf)
	require.NoError(t, err)
	return dbx
}
