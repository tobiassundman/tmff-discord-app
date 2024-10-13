package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"log"
	"time"
	"tmff-discord-app/internal/app/client"
	"tmff-discord-app/internal/app/config"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/internal/controller"
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
		log.Println("database is up to date")
	} else if err != nil {
		log.Fatalf("could not migrate database: %v", err)
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
	defer discordClient.Close()

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

	fanFactionController := controller.NewFanFaction(conf, playerRepo, gameService, leaderboardService, gameScraper)

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "register-game",
			Description: "There has to be at least two registered participants in the game.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game-link",
					Description: "The link to the game on Board Game Arena, e.g. https://boardgamearena.com/table?table=571581855",
					Required:    true,
				},
			},
		},
		{
			Name:        "add-player",
			Description: "Add a player.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The username of the player.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The BGA ID of the player.",
					Required:    true,
				},
			},
		},
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"register-game": fanFactionController.RegisterGame,
		"add-player":    fanFactionController.AddPlayer,
	}

	err = discordClient.Initialize(commands, commandHandlers)
	if err != nil {
		log.Fatalf("could not initialize discord client: %v", err)
	}

	log.Println("Bot is running. Press CTRL+C to exit.")

	// Await a signal to exit
	select {}
}
