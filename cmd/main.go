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

func main() {
	conf, err := config.ReadConfig("config.yaml")
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	parsedQueryTimeout, err := time.ParseDuration(conf.QueryTimeout)
	if err != nil {
		log.Fatalf("could not parse QUERY_TIMEOUT: %v", err)
	}

	dbx := setupDatabase(conf)

	playerRepo := repository.NewPlayer(dbx, &parsedQueryTimeout)
	seasonRepo := repository.NewSeason(dbx, &parsedQueryTimeout, conf.CurrentSeason)
	gameRepo := repository.NewGame(dbx, &parsedQueryTimeout, conf.CurrentSeason)
	gameService := services.NewGame(playerRepo, gameRepo, seasonRepo, conf.EloKFactor)
	leaderboardService := services.NewLeaderboard(seasonRepo, playerRepo)
	discordClient, err := client.NewDiscord(conf)
	if err != nil {
		log.Fatalf("could not create discord client: %v", err)
	}
	defer discordClient.Close()

	// Install playwright
	err = playwright.Install()
	if err != nil {
		log.Printf("could not install playwright: %v", err)
		return
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer func(pw *playwright.Playwright) {
		stopErr := pw.Stop()
		if stopErr != nil {
			log.Printf("could not stop playwright: %v", stopErr)
		}
	}(pw)
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	pages := services.NewPages(browser, conf.MaxGameAgeDays)
	defer pages.Close()

	gameScraper, err := pages.NewGameScraper()
	if err != nil {
		log.Fatalf("could not create game scraper: %v", err)
	}
	defer gameScraper.Close()

	fanFactionController := controller.NewFanFaction(conf, playerRepo, gameService, leaderboardService, gameScraper)
	commands, commandHandlers := fanFactionController.FanFactionCommands()

	err = discordClient.Initialize(commands, commandHandlers)
	if err != nil {
		log.Fatalf("could not initialize discord client: %v", err)
	}

	log.Println("Bot is running. Press CTRL+C to exit.")

	// Await a signal to exit
	select {}
}

func setupDatabase(conf *config.Config) *sqlx.DB {
	db, err := sql.Open("sqlite3", conf.DBFile)
	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}
	dbx := sqlx.NewDb(db, "sqlite3")

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign key constraints:", err)
	}

	err = database.Migrate(db, "./db/migrations")
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("database is up to date")
	} else if err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	return dbx
}
