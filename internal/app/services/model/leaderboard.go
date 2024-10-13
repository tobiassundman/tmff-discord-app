package model

type LeaderboardEntry struct {
	PlayerID    string
	PlayerName  string
	Elo         int
	GamesPlayed int
}
