package model

import "fmt"

type Leaderboard struct {
	Entries []*LeaderboardEntry
}

type LeaderboardEntry struct {
	PlayerID    string
	PlayerName  string
	Elo         int
	GamesPlayed int
}

func (l *Leaderboard) String() string {
	var output string
	output += fmt.Sprintf("%s\n", "Leaderboard")
	header := fmt.Sprintf("%-20s %10s %15s\n", "Player Name", "Elo", "Games Played")
	output += header
	output += fmt.Sprintf("%s\n", "-----------------------------------------------")
	for _, entry := range l.Entries {
		output += fmt.Sprintf("%-20s %10d %15d\n", entry.PlayerName, entry.Elo, entry.GamesPlayed)
	}
	return output
}
