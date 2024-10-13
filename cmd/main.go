package main

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"log"
	"time"
	"tmff-discord-app/internal/app/config"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/pkg/database"
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
		log.Fatalf("database is up to date")
	}
	if err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	// TODO: Use the repository and services
	playerRepo := repository.NewPlayer(dbx, &parsedQueryTimeout)
	seasonRepo := repository.NewSeason(dbx, &parsedQueryTimeout, conf.CurrentSeason)
	gameRepo := repository.NewGame(dbx, &parsedQueryTimeout, conf.CurrentSeason)
	gameService := services.NewGame(playerRepo, gameRepo, seasonRepo, conf.EloKFactor)
	fmt.Println(gameService)

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

	pages := services.NewPages(browser, conf.MaxGameAgeDays)
	defer pages.Close()

	gameScraper, err := pages.NewGameScraper()
	if err != nil {
		log.Fatalf("could not create game scraper: %v", err)
	}
	defer gameScraper.Close()
}
