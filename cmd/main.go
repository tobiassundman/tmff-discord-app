package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/pkg/database"
	"tmff-discord-app/pkg/environment"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
)

func main() {
	var (
		queryTimeoutSeconds  = environment.GetEnvOrDefault("QUERY_TIMEOUT", "5s")
		databaseFile         = environment.GetEnvOrDefault("DB_FILE", ":memory:")
		maxGameAgeDaysString = environment.GetEnvOrDefault("MAX_GAME_AGE_DAYS", "60")
		currentSeason        = environment.GetEnvOrDefault("CURRENT_SEASON", "First Fan Faction Season")
	)

	parsedQueryTimeout, err := time.ParseDuration(queryTimeoutSeconds)
	if err != nil {
		log.Fatalf("could not parse QUERY_TIMEOUT: %v", err)
	}

	maxGameAgeDays, err := strconv.Atoi(maxGameAgeDaysString)
	if err != nil {
		log.Fatalf("could not parse MAX_GAME_AGE_DAYS: %v", err)
	}

	db, err := sql.Open("sqlite3", databaseFile)
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
		log.Fatalf("database is up to date")
	}
	if err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	// TODO: Use the repository and services
	playerRepo := repository.NewPlayer(dbx, &parsedQueryTimeout)
	seasonRepo := repository.NewSeason(dbx, &parsedQueryTimeout, currentSeason)
	gameRepo := repository.NewGame(dbx, &parsedQueryTimeout, currentSeason)
	fmt.Println(playerRepo, seasonRepo, gameRepo)

	err = playwright.Install()
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	pages := services.NewPages(browser, maxGameAgeDays)
	defer pages.Close()

	gameScraper, err := pages.NewGameScraper()
	if err != nil {
		log.Fatalf("could not create game scraper: %v", err)
	}
	defer gameScraper.Close()
}
