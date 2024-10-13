package main

import (
	"database/sql"
	"log"
	"time"
	"tmff-discord-app/internal/app/client"
	"tmff-discord-app/internal/app/config"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/internal/controller"
	"tmff-discord-app/pkg/database"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
)

//nolint:funlen // This is fine :)
func main() {
	conf, err := config.ReadConfig("config.yaml")
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	parsedQueryTimeout, err := time.ParseDuration(conf.QueryTimeout)
	if err != nil {
		log.Fatalf("could not parse QUERY_TIMEOUT: %v", err)
	}

	dbx, err := setupDatabase(conf)
	if err != nil {
		log.Printf("could not setup database: %v", err)
		return
	}

	playerRepo := repository.NewPlayer(dbx, &parsedQueryTimeout)
	seasonRepo := repository.NewSeason(dbx, &parsedQueryTimeout, conf.CurrentSeason)
	gameRepo := repository.NewGame(dbx, &parsedQueryTimeout, conf.CurrentSeason)
	gameService := services.NewGame(playerRepo, gameRepo, seasonRepo, conf.EloKFactor)
	leaderboardService := services.NewLeaderboard(seasonRepo, playerRepo)
	discordClient, err := client.NewDiscord(conf)
	if err != nil {
		log.Fatalf("could not create discord client: %v", err)
	}
	defer func(discordClient *client.Discord) {
		closeErr := discordClient.Close()
		if closeErr != nil {
			log.Printf("could not close discord client: %v", closeErr)
		}
	}(discordClient)

	// Install playwright
	err = playwright.Install()
	if err != nil {
		log.Printf("could not install playwright: %v", err)
		return
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Printf("could not start playwright: %v", err)
		return
	}
	defer func(pw *playwright.Playwright) {
		stopErr := pw.Stop()
		if stopErr != nil {
			log.Printf("could not stop playwright: %v", stopErr)
		}
	}(pw)
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Printf("could not launch browser: %v", err)
		return
	}
	defer func(browser playwright.Browser) {
		closeErr := browser.Close()
		if closeErr != nil {
			log.Printf("could not close browser: %v", closeErr)
		}
	}(browser)

	pages := services.NewPages(browser, conf.MaxGameAgeDays)
	defer func(pages *services.Pages) {
		closeErr := pages.Close()
		if closeErr != nil {
			log.Printf("could not close pages: %v", closeErr)
		}
	}(pages)

	gameScraper, err := pages.NewGameScraper()
	if err != nil {
		log.Printf("could not create game scraper: %v", err)
		return
	}
	defer func(gameScraper *services.GameScraper) {
		closeErr := gameScraper.Close()
		if closeErr != nil {
			log.Printf("could not close game scraper: %v", closeErr)
		}
	}(gameScraper)

	fanFactionController := controller.NewFanFaction(conf, playerRepo, gameService, leaderboardService, gameScraper)
	commands, commandHandlers := fanFactionController.FanFactionCommands()

	err = discordClient.Initialize(commands, commandHandlers)
	if err != nil {
		log.Printf("could not initialize discord client: %v", err)
		return
	}

	log.Println("Bot is running. Press CTRL+C to exit.")

	// Await a signal to exit
	select {}
}

func setupDatabase(conf *config.Config) (*sqlx.DB, error) {
	db, err := sql.Open("sqlite3", conf.DBFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open database")
	}
	dbx := sqlx.NewDb(db, "sqlite3")

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, errors.Wrap(err, "could not enable foreign key constraints")
	}

	err = database.Migrate(db, "./db/migrations")
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("database is up to date")
	} else if err != nil {
		return nil, errors.Wrap(err, "could not migrate database")
	}

	return dbx, nil
}
