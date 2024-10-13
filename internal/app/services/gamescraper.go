package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tmff-discord-app/internal/app/services/model"

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

func (gs *GameScraper) ExtractGameOutcome(inputURL string) (*model.GameOutcome, error) {
	tableID, err := getTableID(inputURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get table ID from URL")
	}
	gameURL := fmt.Sprintf("https://en.boardgamearena.com/table?table=%s", tableID)
	if _, err = gs.page.Goto(gameURL); err != nil {
		return nil, err
	}

	exists, err := gs.tableExists()
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if table exists")
	}
	if !exists {
		return nil, errors.New("table does not exist")
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

	outcome := &model.GameOutcome{
		ID:                tableID,
		Players:           playerResults,
		FanFactionSetting: model.FanFactionSettingFromString(fanFactionSetting),
		CreationTime:      creationTime,
	}
	err = outcome.Validate(gs.maxGameAgeDays)
	if err != nil {
		return nil, errors.Wrap(err, "game outcome is invalid")
	}
	return outcome, nil
}

func getTableID(gameURL string) (string, error) {
	parsedURL, err := url.Parse(gameURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse URL")
	}
	queryParams := parsedURL.Query()
	id := queryParams.Get("table")
	if id == "" {
		return "", errors.New("table ID not found in URL")
	}
	_, err = strconv.Atoi(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert table ID to int")
	}
	return id, nil
}

func (gs *GameScraper) getFanFactionSetting() (string, error) {
	selectOption := gs.page.Locator("#mob_gameoption_108_input option[selected='selected']")
	selectedText, err := selectOption.TextContent()
	if err != nil {
		return "", err
	}
	return selectedText, nil
}

func (gs *GameScraper) getPlayerResults() ([]*model.PlayerResult, error) {
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
	now := time.Now()
	if err != nil {
		return &now, err
	}
	return &parsedTime, nil
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

func extractPlayers(input string) ([]*model.PlayerResult, error) {
	re := regexp.MustCompile(`\d+°\s+([^\(]+)\s+\((\d+)\s+pts\)`)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) != 4 {
		return nil, errors.New("invalid number of players")
	}
	var players []*model.PlayerResult
	for _, match := range matches {
		if len(match) > 2 {
			score, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, err
			}
			players = append(players, &model.PlayerResult{
				Name:  match[1],
				Score: score,
			})
		}
	}
	return players, nil
}

func (gs *GameScraper) tableExists() (bool, error) {
	textLocator := gs.page.Locator("text=Table not found")
	count, err := textLocator.Count()
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	return true, nil
}

func (gs *GameScraper) Close() error {
	return gs.page.Close()
}
