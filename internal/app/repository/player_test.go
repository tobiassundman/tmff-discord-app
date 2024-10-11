package repository_test

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"tmff-discord-app/internal/app/repository"
)

func TestPlayer(t *testing.T) {
	t.Parallel()
	t.Run("Test happy case", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		player, err := playerRepo.GetPlayer("Test Player1")
		require.NoError(t, err)
		assert.Equal(t, "Test Player1", player.Name)
		assert.Equal(t, "1", player.BGAID)
	})

	t.Run("Test insert already exists", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Test Player1", "1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed: players.name")
	})

	t.Run("Test get doesn't exist", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		_, err := playerRepo.GetPlayer("Test Player1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrPlayerNotFound))
	})
}
