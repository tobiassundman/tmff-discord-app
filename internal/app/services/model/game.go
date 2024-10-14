package model

type Player struct {
	Name string
	ID   string
}

type MatchResult struct {
}

type PlayerEloResult struct {
	Name      string
	ID        int
	Score     int
	EloBefore int
	EloChange int
}
