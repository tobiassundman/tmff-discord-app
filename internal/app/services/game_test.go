package services_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/internal/app/services/model"
)

func TestRegisterGame(t *testing.T) {
	t.Parallel()
	t.Run("Register a new game - all players present", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		gameService := services.NewGame(playerRepo, gameRepo, seasonRepo)

		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 2", "2")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 3", "3")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		currentTime := time.Now()
		gameOutcome := &model.GameOutcome{
			ID: "1",
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
		players, err := gameService.RegisterGame(gameOutcome)
		require.NoError(t, err)
		assert.Len(t, players, 4)
		assert.Equal(t, "Player 4", players[0].Name)
		assert.Equal(t, "4", players[0].ID)
		assert.Equal(t, 400, players[0].Score)
		assert.Equal(t, 1000, players[0].EloBefore)
		assert.Equal(t, 15, players[0].EloChange)
		assert.Equal(t, "Player 3", players[1].Name)
		assert.Equal(t, "3", players[1].ID)
		assert.Equal(t, 300, players[1].Score)
		assert.Equal(t, 1000, players[1].EloBefore)
		assert.Equal(t, 5, players[1].EloChange)
		assert.Equal(t, "Player 2", players[2].Name)
		assert.Equal(t, "2", players[2].ID)
		assert.Equal(t, 200, players[2].Score)
		assert.Equal(t, 1000, players[2].EloBefore)
		assert.Equal(t, -5, players[2].EloChange)
		assert.Equal(t, "Player 1", players[3].Name)
		assert.Equal(t, "1", players[3].ID)
		assert.Equal(t, 100, players[3].Score)
		assert.Equal(t, 1000, players[3].EloBefore)
		assert.Equal(t, -15, players[3].EloChange)
		for _, player := range players {
			fmt.Printf("Player: %s, Score: %d, EloBefore: %d, EloChange: %d\n", player.Name, player.Score, player.EloBefore, player.EloChange)
		}
	})
	t.Run("Register a new game - three players present", func(t *testing.T) {
		t.Parallel()
		dbx := newMigratedSQLiteDB(t)
		queryTimeout := 2 * time.Second
		gameRepo := repository.NewGame(dbx, &queryTimeout, "First Fan Faction Season")
		playerRepo := repository.NewPlayer(dbx, &queryTimeout)
		seasonRepo := repository.NewSeason(dbx, &queryTimeout, "First Fan Faction Season")
		gameService := services.NewGame(playerRepo, gameRepo, seasonRepo)

		err := playerRepo.InsertPlayer("Player 1", "1")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 2", "2")
		require.NoError(t, err)
		err = playerRepo.InsertPlayer("Player 4", "4")
		require.NoError(t, err)

		currentTime := time.Now()
		gameOutcome := &model.GameOutcome{
			ID: "1",
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
		players, err := gameService.RegisterGame(gameOutcome)
		require.NoError(t, err)
		assert.Len(t, players, 3)
		assert.Equal(t, "Player 4", players[0].Name)
		assert.Equal(t, "4", players[0].ID)
		assert.Equal(t, 400, players[0].Score)
		assert.Equal(t, 1000, players[0].EloBefore)
		assert.Equal(t, 10, players[0].EloChange)
		assert.Equal(t, "Player 2", players[1].Name)
		assert.Equal(t, "2", players[1].ID)
		assert.Equal(t, 200, players[1].Score)
		assert.Equal(t, 1000, players[1].EloBefore)
		assert.Equal(t, 0, players[1].EloChange)
		assert.Equal(t, "Player 1", players[2].Name)
		assert.Equal(t, "1", players[2].ID)
		assert.Equal(t, 100, players[2].Score)
		assert.Equal(t, 1000, players[2].EloBefore)
		assert.Equal(t, -10, players[2].EloChange)
		for _, player := range players {
			fmt.Printf("Player: %s, Score: %d, EloBefore: %d, EloChange: %d\n", player.Name, player.Score, player.EloBefore, player.EloChange)
		}
	})
}
