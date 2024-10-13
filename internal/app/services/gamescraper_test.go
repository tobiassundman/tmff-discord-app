package services_test

import (
	"testing"
	"time"
	"tmff-discord-app/internal/app/services"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	err := playwright.Install()
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestExtractGameOutcome(t *testing.T) {
	t.Parallel()
	t.Run("friendly mode, correct settings", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		gameOutcome, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=572461868")
		require.NoError(t, err)

		assert.Equal(t, "On - no Fire & Ice", string(gameOutcome.FanFactionSetting))
		assert.Equal(t, "2024-10-07T21:01:00Z", gameOutcome.CreationTime.Format(time.RFC3339))
		assert.Len(t, gameOutcome.Players, 4)
		assert.Equal(t, "Stahlbrötchen", gameOutcome.Players[0].Name)
		assert.Equal(t, 148, gameOutcome.Players[0].Score)
		assert.Equal(t, "deragned", gameOutcome.Players[1].Name)
		assert.Equal(t, 146, gameOutcome.Players[1].Score)
		assert.Equal(t, "skoomymooms", gameOutcome.Players[2].Name)
		assert.Equal(t, 133, gameOutcome.Players[2].Score)
		assert.Equal(t, "Zaarito", gameOutcome.Players[3].Name)
		assert.Equal(t, 100, gameOutcome.Players[3].Score)
	})
	t.Run("turn based, correct settings", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		gameOutcome, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=559705570")
		require.NoError(t, err)

		assert.Equal(t, "On - with Fire & Ice", string(gameOutcome.FanFactionSetting))
		assert.Equal(t, "2024-09-07T15:43:00Z", gameOutcome.CreationTime.Format(time.RFC3339))
		assert.Len(t, gameOutcome.Players, 4)
		assert.Equal(t, "ymse", gameOutcome.Players[0].Name)
		assert.Equal(t, 152, gameOutcome.Players[0].Score)
		assert.Equal(t, "vonbrot", gameOutcome.Players[1].Name)
		assert.Equal(t, 149, gameOutcome.Players[1].Score)
		assert.Equal(t, "korkje", gameOutcome.Players[2].Name)
		assert.Equal(t, 132, gameOutcome.Players[2].Score)
		assert.Equal(t, "hugnad", gameOutcome.Players[3].Name)
		assert.Equal(t, 130, gameOutcome.Players[3].Score)
	})

	t.Run("wrong game - yahtzee", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=544240084")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "game name is not Terra Mystica")
	})

	t.Run("fan factions not enabled", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=570819150")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fan factions are not enabled")
	})

	t.Run("wrong player count", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=557774225")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid number of players")
	})
	t.Run("abandoned game", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=555675245")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "game outcome is invalid: player 3 has no score")
	})

	t.Run("table does not exist", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=555555555555555")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "table does not exist")
	})

	t.Run("not an url", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get table ID from URL")
	})

	t.Run("url lacks table id", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		_, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get table ID from URL")
	})

	t.Run("more query parameters work", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		gameOutcome, err := gameScraper.ExtractGameOutcome("https://boardgamearena.com/table?table=572461868&a=b")
		require.NoError(t, err)
		assert.Len(t, gameOutcome.Players, 4)
	})

	t.Run("other language works", func(t *testing.T) {
		t.Parallel()
		gameScraper := createGameScraper(t)
		defer gameScraper.Close()

		gameOutcome, err := gameScraper.ExtractGameOutcome("sv.boardgamearena.com//table?table=572461868&a=b")
		require.NoError(t, err)
		assert.Len(t, gameOutcome.Players, 4)
	})
}

func createGameScraper(t *testing.T) *services.GameScraper {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() {
		stopErr := pw.Stop()
		if stopErr != nil {
			t.Logf("could not stop playwright: %v", stopErr)
		}
	})

	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() {
		browser.Close()
	})

	pages := services.NewPages(browser, 100000)
	t.Cleanup(func() {
		pages.Close()
	})

	gameScraper, err := pages.NewGameScraper()
	require.NoError(t, err)
	return gameScraper
}
