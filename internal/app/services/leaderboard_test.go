package services_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/internal/app/services/model"
)

func TestGetLeaderboard(t *testing.T) {
	t.Parallel()
	t.Run("Test happy case", func(t *testing.T) {
		t.Parallel()

		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		gameService := services.NewGame(playerRepo, gameRepo, seasonRepo, K)
		leaderboardService := services.NewLeaderboard(seasonRepo, playerRepo)

		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 2", "2")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 3", "3")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		currentTime := time.Now()
		gameOutcome1 := &model.GameOutcome{
			ID: "1",
			Players: []*model.PlayerResult{
				{
					Name:  "Unregistered player 1",
					Score: 100,
				},
				{
					Name:  "Unregistered player 2",
					Score: 200,
				},
				{
					Name:  "Player 3",
					Score: 300,
				},
				{
					Name:  "Player 4",
					Score: 400,
				},
			},
			FanFactionSetting: model.On,
			CreationTime:      &currentTime,
		}
		_, err = gameService.RegisterGame(gameOutcome1)
		require.NoError(t, err)

		gameOutcome2 := &model.GameOutcome{
			ID: "2",
			Players: []*model.PlayerResult{
				{
					Name:  "Player 1",
					Score: 100,
				},
				{
					Name:  "Player 2",
					Score: 200,
				},
				{
					Name:  "Player 3",
					Score: 300,
				},
				{
					Name:  "Unregistered Player",
					Score: 400,
				},
			},
			FanFactionSetting: model.OnNoFireAndIce,
			CreationTime:      &currentTime,
		}
		_, err = gameService.RegisterGame(gameOutcome2)
		require.NoError(t, err)

		gameOutcome3 := &model.GameOutcome{
			ID: "3",
			Players: []*model.PlayerResult{
				{
					Name:  "Player 1",
					Score: 100,
				},
				{
					Name:  "Player 2",
					Score: 200,
				},
				{
					Name:  "Player 3",
					Score: 300,
				},
				{
					Name:  "Player 4",
					Score: 400,
				},
			},
			FanFactionSetting: model.OnNoFireAndIce,
			CreationTime:      &currentTime,
		}
		_, err = gameService.RegisterGame(gameOutcome3)
		require.NoError(t, err)

		gameOutcome4 := &model.GameOutcome{
			ID: "4",
			Players: []*model.PlayerResult{
				{
					Name:  "Player 1",
					Score: 100,
				},
				{
					Name:  "Player 2",
					Score: 200,
				},
				{
					Name:  "Player 3",
					Score: 300,
				},
				{
					Name:  "Player 4",
					Score: 400,
				},
			},
			FanFactionSetting: model.OnNoFireAndIce,
			CreationTime:      &currentTime,
		}
		_, err = gameService.RegisterGame(gameOutcome4)
		require.NoError(t, err)

		leaderboard, err := leaderboardService.GetLeaderboard()
		require.NoError(t, err)
		assert.Len(t, leaderboard, 4)
		assert.Equal(t, "Player 4", leaderboard[0].PlayerName)
		assert.Equal(t, "4", leaderboard[0].PlayerID)
		assert.Equal(t, 1069, leaderboard[0].Elo)
		assert.Equal(t, 3, leaderboard[0].GamesPlayed)
		assert.Equal(t, "Player 3", leaderboard[1].PlayerName)
		assert.Equal(t, "3", leaderboard[1].PlayerID)
		assert.Equal(t, 1028, leaderboard[1].Elo)
		assert.Equal(t, 4, leaderboard[1].GamesPlayed)
		assert.Equal(t, "Player 2", leaderboard[2].PlayerName)
		assert.Equal(t, "2", leaderboard[2].PlayerID)
		assert.Equal(t, 980, leaderboard[2].Elo)
		assert.Equal(t, 3, leaderboard[2].GamesPlayed)
		assert.Equal(t, "Player 1", leaderboard[3].PlayerName)
		assert.Equal(t, "1", leaderboard[3].PlayerID)
		assert.Equal(t, 923, leaderboard[3].Elo)
		assert.Equal(t, 3, leaderboard[3].GamesPlayed)
	})
}
