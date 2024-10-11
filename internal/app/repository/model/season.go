package model

import "time"

const StartElo = 1000

type SeasonParticipant struct {
	ID          int       `db:"id"`
	SeasonName  string    `db:"season_name"`
	PlayerID    string    `db:"player_id"`
	Elo         int       `db:"elo"`
	GamesPlayed int       `db:"games_played"`
	CreatedAt   time.Time `db:"created_at"`
}
