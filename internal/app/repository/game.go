package repository

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
	"time"
	"tmff-discord-app/internal/app/repository/model"
)

var (
	ErrGameNotFound = errors.New("game not found")
)

const (
	insertGameQuery = `
		INSERT INTO games (bga_id, season_name) 
		VALUES ($1, $2)`
	insertGameParticipantQuery = `
		INSERT INTO game_participants (game_id, player_id, score, elo_change, elo_before) 
		VALUES ($1, $2, $3, $4, $5)`
	selectParticipantsQuery = `
		SELECT 
			id, 
			game_id, 
			player_id, 
			score, 
			elo_change, 
			elo_before, 
			created_at 
		FROM 
			game_participants 
		WHERE 
			game_id = $1`
	selectGameQuery = `SELECT bga_id, season_name, created_at FROM games WHERE bga_id = $1`
)

type Game struct {
	db           *sqlx.DB
	queryTimeout *time.Duration
	seasonName   string
}

func NewGame(db *sqlx.DB, queryTimeout *time.Duration, seasonName string) *Game {
	return &Game{
		db:           db,
		queryTimeout: queryTimeout,
		seasonName:   seasonName,
	}
}

func (r *Game) CreateGameWithParticipants(gameID string, participants []*model.GameParticipant) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Insert game
	_, err = tx.Exec(insertGameQuery, gameID, r.seasonName)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return errors.New("season does not exist")
		}
		return errors.Wrap(err, "failed to insert game")
	}

	// Insert game participants
	for _, participant := range participants {
		_, err = tx.Exec(insertGameParticipantQuery, gameID, participant.PlayerID, participant.Score, participant.EloChange, participant.EloBefore)
		if err != nil {
			if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
				return errors.New("player does not exist")
			}
			return errors.Wrap(err, "failed to insert game participant")
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (r *Game) GetGameWithParticipants(gameID string) (*model.GameWithParticipants, error) {
	var game model.Game
	var participants []model.GameParticipant

	// Get game details
	err := r.db.Get(&game, selectGameQuery, gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGameNotFound
		}
		return nil, errors.Wrap(err, "failed to query game")
	}

	// Get game participants
	rows, err := r.db.Queryx(selectParticipantsQuery, gameID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query game participants")
	}
	defer rows.Close()

	for rows.Next() {
		var participant model.GameParticipant
		scanErr := rows.StructScan(&participant)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "failed to scan participant")
		}
		participants = append(participants, participant)
	}

	gameWithParticipants := &model.GameWithParticipants{
		GameID:       game.BGAID,
		SeasonName:   game.SeasonName,
		CreatedAt:    game.CreatedAt,
		Participants: participants,
	}

	return gameWithParticipants, nil
}
