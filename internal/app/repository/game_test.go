package repository_test

import (
	"testing"
	"time"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/repository/model"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGame(t *testing.T) {
	t.Parallel()
	t.Run("Test happy case", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		// Create players
		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 2", "2")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 3", "3")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		gameID := "1"
		participants := []*model.GameParticipant{
			{
				PlayerID:  1,
				Score:     110,
				EloChange: 20,
				EloBefore: 1000,
			},
			{
				PlayerID:  2,
				Score:     130,
				EloChange: -20,
				EloBefore: 1000,
			},
			{
				PlayerID:  3,
				Score:     100,
				EloChange: -10,
				EloBefore: 1000,
			},
			{
				PlayerID:  4,
				Score:     140,
				EloChange: 50,
				EloBefore: 1000,
			},
		}
		err = gameRepo.CreateGameWithParticipants(gameID, participants)
		require.NoError(t, err)
		game, err := gameRepo.GetGameWithParticipants(gameID)
		require.NoError(t, err)

		assert.Equal(t, gameID, game.GameID)
		assert.Equal(t, "First Fan Faction Season", game.SeasonName)
		assert.Len(t, game.Participants, 4)

		assert.Equal(t, 1, game.Participants[0].PlayerID)
		assert.Equal(t, 110, game.Participants[0].Score)
		assert.Equal(t, 20, game.Participants[0].EloChange)
		assert.Equal(t, 1000, game.Participants[0].EloBefore)
		assert.Equal(t, 2, game.Participants[1].PlayerID)
		assert.Equal(t, 130, game.Participants[1].Score)
		assert.Equal(t, -20, game.Participants[1].EloChange)
		assert.Equal(t, 1000, game.Participants[1].EloBefore)
		assert.Equal(t, 3, game.Participants[2].PlayerID)
		assert.Equal(t, 100, game.Participants[2].Score)
		assert.Equal(t, -10, game.Participants[2].EloChange)
		assert.Equal(t, 1000, game.Participants[2].EloBefore)
		assert.Equal(t, 4, game.Participants[3].PlayerID)
		assert.Equal(t, 140, game.Participants[3].Score)
		assert.Equal(t, 50, game.Participants[3].EloChange)
		assert.Equal(t, 1000, game.Participants[3].EloBefore)
	})

	t.Run("Player doesn't exist", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		// Create players
		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 3", "3")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		gameID := "1"
		participants := []*model.GameParticipant{
			{
				PlayerID:  1,
				Score:     110,
				EloChange: 20,
				EloBefore: 1000,
			},
			{
				PlayerID:  2,
				Score:     130,
				EloChange: -20,
				EloBefore: 1000,
			},
			{
				PlayerID:  3,
				Score:     100,
				EloChange: -10,
				EloBefore: 1000,
			},
			{
				PlayerID:  4,
				Score:     140,
				EloChange: 50,
				EloBefore: 1000,
			},
		}
		err = gameRepo.CreateGameWithParticipants(gameID, participants)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "player does not exist")
		_, err = gameRepo.GetGameWithParticipants(gameID)
		require.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrGameNotFound))
	})

	t.Run("Season doesn't exist", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "Second Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		// Create players
		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 2", "2")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 3", "3")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		gameID := "1"
		participants := []*model.GameParticipant{
			{
				PlayerID:  1,
				Score:     110,
				EloChange: 20,
				EloBefore: 1000,
			},
			{
				PlayerID:  2,
				Score:     130,
				EloChange: -20,
				EloBefore: 1000,
			},
			{
				PlayerID:  3,
				Score:     100,
				EloChange: -10,
				EloBefore: 1000,
			},
			{
				PlayerID:  4,
				Score:     140,
				EloChange: 50,
				EloBefore: 1000,
			},
		}
		err = gameRepo.CreateGameWithParticipants(gameID, participants)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "season does not exist")
		_, err = gameRepo.GetGameWithParticipants(gameID)
		require.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrGameNotFound))
	})

	t.Run("Game already exists", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)

		// Create players
		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 2", "2")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 3", "3")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		gameID := "1"
		participants := []*model.GameParticipant{
			{
				PlayerID:  1,
				Score:     110,
				EloChange: 20,
				EloBefore: 1000,
			},
			{
				PlayerID:  2,
				Score:     130,
				EloChange: -20,
				EloBefore: 1000,
			},
			{
				PlayerID:  3,
				Score:     100,
				EloChange: -10,
				EloBefore: 1000,
			},
			{
				PlayerID:  4,
				Score:     140,
				EloChange: 50,
				EloBefore: 1000,
			},
		}
		err = gameRepo.CreateGameWithParticipants(gameID, participants)
		require.NoError(t, err)
		err = gameRepo.CreateGameWithParticipants(gameID, participants)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "game is already registered")
	})
}
