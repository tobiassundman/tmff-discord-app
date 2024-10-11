package database

import (
	"fmt"
	"time"
	"tmff-discord-app/pkg/retry"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

// UserDatabaseConnection creates a connection to the user database and retries ping until it succeeds or times out.
func UserDatabaseConnection(host, port, user, password, name string) (*sqlx.DB, error) {
	configString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name,
	)
	connectionString, err := pgx.ParseConnectionString(configString)
	if err != nil {
		return nil, err
	}
	connection := stdlib.OpenDB(connectionString)

	sqlDB := sqlx.NewDb(connection, "pgx")

	err = retry.Retry(time.Minute, func() error {
		return sqlDB.Ping()
	})
	return sqlDB, err
}
