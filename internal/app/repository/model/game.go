package model

import "time"

type GameWithParticipants struct {
	GameID       string
	SeasonName   string
	CreatedAt    time.Time
	Participants []GameParticipant
}

type Game struct {
	BGAID      string    `db:"bga_id"`
	SeasonName string    `db:"season_name"`
	CreatedAt  time.Time `db:"created_at"`
}

type GameParticipant struct {
	ID        int       `db:"id"`
	GameID    string    `db:"game_id"`
	PlayerID  string    `db:"player_id"`
	Score     int       `db:"score"`
	EloChange int       `db:"elo_change"`
	EloBefore int       `db:"elo_before"`
	CreatedAt time.Time `db:"created_at"`
}
