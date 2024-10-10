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
	Players           []*PlayerResult
	FanFactionSetting FanFactionSetting
	CreationTime      *time.Time
}

func (g *GameOutcome) String() string {
	var output string
	for _, player := range g.Players {
		output += fmt.Sprintf("PlayerResult: %s, Score: %d\n", player.Name, player.Score)
	}
	output += fmt.Sprintf("Fan Faction Setting: %s\n", g.FanFactionSetting)
	output += fmt.Sprintf("Creation Time: %s\n", g.CreationTime)
	return output
}

func (o *GameOutcome) Validate() error {
	for i, player := range o.Players {
		if player.Name == "" {
			return fmt.Errorf("player %d has no name", i)
		}
		if player.Score <= 0 {
			return fmt.Errorf("player %d has no score", i)
		}
	}
	if o.FanFactionSetting != On && o.FanFactionSetting != OnNoFireAndIce {
		return errors.New("fan factions are not enabled")
	}
	if o.CreationTime.Before(time.Now().Add(-60 * time.Hour * 24)) {
		return errors.New("game is too old (more than 60 days)")
	}
	return nil
}
