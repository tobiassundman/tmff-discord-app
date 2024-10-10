package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	model2 "tmff-discord-app/internal/app/services/model"

	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
)

type GameScraper struct {
	page           playwright.Page
	maxGameAgeDays int
}

func newGameScraper(page playwright.Page, maxGameAgeDays int) *GameScraper {
	return &GameScraper{
		page:           page,
		maxGameAgeDays: maxGameAgeDays,
	}
}

func (gs *GameScraper) ExtractGameOutcome(inputURL string) (*model2.GameOutcome, error) {
	tableID, err := getTableID(inputURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get table ID from URL")
	}
	gameURL := fmt.Sprintf("https://en.boardgamearena.com/table?table=%d", tableID)
	if _, err = gs.page.Goto(gameURL); err != nil {
		return nil, err
	}

	err = gs.assertIsTerraMystica()
	if err != nil {
		return nil, errors.Wrap(err, "game is not Terra Mystica")
	}

	fanFactionSetting, err := gs.getFanFactionSetting()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get fan faction setting")
	}
	playerResults, err := gs.getPlayerResults()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get player results")
	}

	creationTime, err := gs.getCreationTime()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get creation time")
	}

	outcome := &model2.GameOutcome{
		Players:           playerResults,
		FanFactionSetting: model2.FanFactionSettingFromString(fanFactionSetting),
		CreationTime:      creationTime,
	}
	err = outcome.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "game outcome is invalid")
	}
	return outcome, nil
}

func getTableID(gameURL string) (int, error) {
	parsedURL, err := url.Parse(gameURL)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse URL")
	}
	queryParams := parsedURL.Query()
	id := queryParams.Get("table")
	if id == "" {
		return 0, errors.New("table ID not found in URL")
	}
	tableID, err := strconv.Atoi(id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert table ID to int")
	}
	return tableID, nil
}

func (gs *GameScraper) getFanFactionSetting() (string, error) {
	selectOption := gs.page.Locator("#mob_gameoption_108_input option[selected='selected']")
	selectedText, err := selectOption.TextContent()
	if err != nil {
		return "", err
	}
	return selectedText, nil
}

func (gs *GameScraper) getPlayerResults() ([]*model2.PlayerResult, error) {
	resultElement := gs.page.Locator(`meta[property="og:description"][content*="1°"]`)
	results, err := resultElement.GetAttribute("content")
	if err != nil {
		return nil, errors.Wrap(err, "could not get content attribute")
	}
	return extractPlayers(results)
}

func (gs *GameScraper) getCreationTime() (*time.Time, error) {
	divElement := gs.page.Locator(`#creationtime`)
	textContent, err := divElement.TextContent()
	if err != nil {
		return nil, err
	}
	// Define the layout matching the date string format
	layout := "Created 01/02/2006 at 15:04"

	// Parse the date string into a time.Time object
	parsedTime, err := time.Parse(layout, textContent)
	return &parsedTime, err
}

func (gs *GameScraper) assertIsTerraMystica() error {
	titleElement := gs.page.Locator(`meta[property="og:title"]`)
	title, err := titleElement.GetAttribute("content")
	if err != nil {
		return errors.Wrap(err, "could not get content attribute")
	}
	if strings.Contains(title, ":") {
		gameTitle := strings.Split(title, ":")[0]
		if gameTitle == "Terra Mystica" {
			return nil
		}
	}
	return errors.New("game name is not Terra Mystica")
}

func extractPlayers(input string) ([]*model2.PlayerResult, error) {
	re := regexp.MustCompile(`\d+°\s+(\w+)\s+\((\d+)\s+pts\)`)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) != 4 {
		return nil, errors.New("invalid number of players")
	}
	var players []*model2.PlayerResult
	for _, match := range matches {
		if len(match) > 2 {
			score, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, err
			}
			players = append(players, &model2.PlayerResult{
				Name:  match[1],
				Score: score,
			})
		}
	}
	return players, nil
}

func (gs *GameScraper) Close() error {
	return gs.page.Close()
}
