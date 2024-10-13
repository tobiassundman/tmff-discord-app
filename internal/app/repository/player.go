package repository

import (
	"database/sql"
	"github.com/pkg/errors"
	"time"
	"tmff-discord-app/internal/app/repository/model"

	"github.com/jmoiron/sqlx"
)

var (
	ErrPlayerNotFound = errors.New("player doesn't exist")
)

const (
	getPlayerQuery     = `SELECT name, bga_id, created_at FROM players WHERE name = $1`
	insertPlayerQuery  = `INSERT INTO players (name, bga_id) VALUES ($1, $2)`
	getAllPlayersQuery = `SELECT name, bga_id, created_at FROM players`
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPlayerNotFound
	}
	return &player, err
}

func (p *Player) GetPlayers() ([]*model.Player, error) {
	var players []*model.Player
	err := p.db.Select(&players, getAllPlayersQuery)
	if err != nil {
		return nil, err
	}
	return players, nil
}

func (p *Player) InsertPlayer(name, bgaID string) error {
	_, err := p.db.Exec(insertPlayerQuery, name, bgaID)
	return err
}
