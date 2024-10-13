package repository_test

import (
	"testing"
	"time"
	"tmff-discord-app/internal/app/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAll(t *testing.T) {
	t.Parallel()
	t.Run("Test Get All", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 1)
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Test Player2", "2")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("2", 2)
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Test Player3", "3")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("3", -3)
		require.NoError(t, err)

		result, err := seasonRepo.GetAll()
		require.NoError(t, err)

		assert.Len(t, result, 3)
		assert.Equal(t, "1", result[1].PlayerID)
		assert.Equal(t, "First Fan Faction Season", result[1].SeasonName)
		assert.Equal(t, 1001, result[1].Elo)
		assert.Equal(t, 1, result[1].GamesPlayed)
		assert.Equal(t, "2", result[0].PlayerID)
		assert.Equal(t, "First Fan Faction Season", result[1].SeasonName)
		assert.Equal(t, 1002, result[0].Elo)
		assert.Equal(t, 1, result[0].GamesPlayed)
		assert.Equal(t, "3", result[2].PlayerID)
		assert.Equal(t, "First Fan Faction Season", result[2].SeasonName)
		assert.Equal(t, 997, result[2].Elo)
	})
}

func TestUpsert(t *testing.T) {
	t.Parallel()
	t.Run("Test Upsert when doesn't exist", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 1)
		require.NoError(t, err)
	})

	t.Run("Test Upsert when exists", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 1)
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 2)
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 3)
		require.NoError(t, err)

		result, err := seasonRepo.GetAll()
		require.NoError(t, err)

		assert.Len(t, result, 1)
		assert.Equal(t, 1006, result[0].Elo)
		assert.Equal(t, 3, result[0].GamesPlayed)
	})

	t.Run("Test Upsert below 0 returns 0 elo", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 1)
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", -10000)
		require.NoError(t, err)

		result, err := seasonRepo.GetAll()
		require.NoError(t, err)

		assert.Len(t, result, 1)
		assert.Equal(t, 0, result[0].Elo)
	})

	t.Run("Test Upsert when player doesn't exist", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")

		_, err := seasonRepo.UpsertSeasonParticipant("1", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "player or season does not exist")
	})

	t.Run("Test Upsert when season doesn't exist", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "Second Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		err := playerRepo.InsertPlayer("Test Player1", "1")
		require.NoError(t, err)
		_, err = seasonRepo.UpsertSeasonParticipant("1", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "player or season does not exist")
	})
}
