package repository

import (
	"time"
	"tmff-discord-app/internal/app/repository/model"

	"github.com/jmoiron/sqlx"
)

const (
	getPlayerQuery    = `SELECT name, elo, bga_id, created_at FROM players WHERE name = $1`
	insertPlayerQuery = `INSERT INTO players (name, bga_id) VALUES ($1, $2)`
	updatePlayerQuery = `UPDATE players SET elo = $1 WHERE name = $2`
	deletePlayerQuery = `DELETE FROM players WHERE name = $1`
)

type Player struct {
	db           *sqlx.DB
	queryTimeout *time.Duration
}

func NewPlayer(db *sqlx.DB, queryTimeout *time.Duration) *Player {
	return &Player{
		db:           db,
		queryTimeout: queryTimeout,
	}
}

func (p *Player) GetPlayer(name string) (*model.Player, error) {
	var player model.Player
	err := p.db.Get(&player, getPlayerQuery, name)
	return &player, err
}

func (p *Player) InsertPlayer(name, bgaID string) error {
	_, err := p.db.Exec(insertPlayerQuery, name, bgaID)
	return err
}

func (p *Player) UpdatePlayer(name string, elo float64) error {
	_, err := p.db.Exec(updatePlayerQuery, elo, name)
	if err != nil {
		return err
	}
	return nil
}

func (p *Player) DeletePlayer(name string) error {
	_, err := p.db.Exec(deletePlayerQuery, name)
	if err != nil {
		return err
	}
	return nil
}
