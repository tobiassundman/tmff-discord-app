package services

import "github.com/playwright-community/playwright-go"

type Pages struct {
	browser        playwright.Browser
	maxGameAgeDays int
}

func NewPages(browser playwright.Browser, maxGameAgeDays int) *Pages {
	return &Pages{
		browser:        browser,
		maxGameAgeDays: maxGameAgeDays,
	}
}

func (p *Pages) NewGameScraper() (*GameScraper, error) {
	page, err := p.browser.NewPage()
	if err != nil {
		return nil, err
	}
	return newGameScraper(page, p.maxGameAgeDays), nil
}

func (p *Pages) Close() error {
	return p.browser.Close()
}
