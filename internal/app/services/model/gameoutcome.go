package model

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type PlayerResult struct {
	Name  string
	Score int
}

type GameOutcome struct {
	ID                string
	Players           []*PlayerResult
	FanFactionSetting FanFactionSetting
	CreationTime      *time.Time
}

func (g *GameOutcome) String() string {
	var output string
	output += fmt.Sprintf("https://boardgamearena.com/table?table=%s\n", g.ID)
	for _, player := range g.Players {
		output += fmt.Sprintf("%s, Score: %d\n", player.Name, player.Score)
	}
	return output
}

func (g *GameOutcome) Validate(maxGameAgeDays int) error {
	for i, player := range g.Players {
		if player.Name == "" {
			return fmt.Errorf("player %d has no name", i)
		}
		if player.Score <= 0 {
			return fmt.Errorf("player %d has no score", i)
		}
	}
	if g.FanFactionSetting != On && g.FanFactionSetting != OnNoFireAndIce {
		return errors.New("fan factions are not enabled")
	}
	if g.CreationTime.Before(time.Now().Add(-time.Duration(maxGameAgeDays) * time.Hour * 24)) {
		return errors.New("game is too old (more than 60 days)")
	}
	return nil
}
