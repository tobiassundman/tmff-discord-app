package repository

import (
	"time"
	"tmff-discord-app/internal/app/repository/model"

	"github.com/jmoiron/sqlx"
)

const (
	getPlayerQuery    = `SELECT name, bga_id, created_at FROM players WHERE name = $1`
	insertPlayerQuery = `INSERT INTO players (name, bga_id) VALUES ($1, $2)`
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
