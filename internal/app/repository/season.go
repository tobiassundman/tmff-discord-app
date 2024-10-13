package repository

import (
	"database/sql"
	"log"
	"strings"
	"time"
	"tmff-discord-app/internal/app/repository/model"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

const (
	getAllSeasonParticipantsQuery = `
		SELECT id, season_name, player_id, elo, games_played, created_at 
		FROM season_participants 
		WHERE season_name = $1 
		ORDER BY elo DESC`
	getSeasonParticipantQuery = `
		SELECT id, season_name, player_id, elo, games_played, created_at 
		FROM season_participants 
		WHERE player_id = $1`
	insertSeasonParticipantQuery = `
		INSERT INTO season_participants(season_name, player_id, elo, games_played) 
		VALUES(:season_name,:player_id,:elo,:games_played)`
	updateSeasonParticipantQuery = `
		UPDATE season_participants 
		SET elo =:elo, games_played =:games_played 
		WHERE id =:id`
)

type Season struct {
	db            *sqlx.DB
	queryTimeout  *time.Duration
	currentSeason string
}

func NewSeason(db *sqlx.DB, queryTimeout *time.Duration, currentSeason string) *Season {
	return &Season{
		db:            db,
		queryTimeout:  queryTimeout,
		currentSeason: currentSeason,
	}
}

func (s *Season) GetAll() ([]*model.SeasonParticipant, error) {
	var participants []*model.SeasonParticipant
	err := s.db.Select(&participants, getAllSeasonParticipantsQuery, s.currentSeason)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

func (s *Season) UpsertSeasonParticipant(playerID string, eloChange int) (*model.SeasonParticipant, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func(tx *sqlx.Tx) {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("failed to rollback transaction: %v", rollbackErr)
		}
	}(tx)

	// Get the current season participant
	var participant model.SeasonParticipant
	err = tx.Get(&participant, getSeasonParticipantQuery, playerID)
	if errors.Is(err, sql.ErrNoRows) {
		// Create the participant
		participant = model.SeasonParticipant{
			SeasonName:  s.currentSeason,
			PlayerID:    playerID,
			Elo:         model.StartElo + eloChange,
			GamesPlayed: 1,
		}
		_, err = tx.NamedExec(insertSeasonParticipantQuery, participant)
		switch {
		case err != nil && strings.Contains(err.Error(), "FOREIGN KEY constraint failed"):
			return nil, errors.New("player or season does not exist")
		case err != nil:
			return nil, err
		}

		commitErr := tx.Commit()
		if commitErr != nil {
			return nil, commitErr
		}
		return &participant, nil
	} else if err != nil {
		return nil, err
	}

	// Update the participant
	// Don't go below 0 elo
	if participant.Elo+eloChange < 0 {
		eloChange = -participant.Elo
	}
	participant.Elo += eloChange
	participant.GamesPlayed++
	_, err = tx.NamedExec(updateSeasonParticipantQuery, participant)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &participant, nil
}
