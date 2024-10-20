package model

import "fmt"

type Leaderboard struct {
	Entries []*LeaderboardEntry
}

type LeaderboardEntry struct {
	PlayerName  string
	Elo         int
	GamesPlayed int
}

func (l *Leaderboard) String() string {
	var output string
	output += fmt.Sprintf("%s\n", "Leaderboard")
	header := fmt.Sprintf("%-4s %-20s %4s %12s\n", "Rank", "Player Name", "Elo", "Games Played")
	output += header
	output += fmt.Sprintf("%s\n", "-------------------------------------------")
	for index, entry := range l.Entries {
		playerNameTruncated := truncateString(entry.PlayerName, 20)
		output += fmt.Sprintf("%-4d %-20s %4d %12d\n", index+1, playerNameTruncated, entry.Elo, entry.GamesPlayed)
	}
	return output
}

func truncateString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}
