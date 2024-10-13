package services

import (
	"sort"
	"tmff-discord-app/internal/app/repository"
	"tmff-discord-app/internal/app/services/model"
)

type Leaderboard struct {
	seasonRepo *repository.Season
	playerRepo *repository.Player
}

func NewLeaderboard(seasonRepo *repository.Season, playerRepo *repository.Player) *Leaderboard {
	return &Leaderboard{
		seasonRepo: seasonRepo,
		playerRepo: playerRepo,
	}
}

func (l *Leaderboard) GetLeaderboard() ([]*model.LeaderboardEntry, error) {
	seasonParticipants, err := l.seasonRepo.GetAll()
	if err != nil {
		return nil, err
	}

	players, err := l.playerRepo.GetPlayers()
	if err != nil {
		return nil, err
	}

	leaderboardEntries := make([]*model.LeaderboardEntry, len(seasonParticipants))

	// sort by elo
	sort.Slice(seasonParticipants, func(i, j int) bool {
		return seasonParticipants[i].Elo > seasonParticipants[j].Elo
	})

	for i, participant := range seasonParticipants {
		var playerName string
		for _, player := range players {
			if player.BGAID == participant.PlayerID {
				playerName = player.Name
				break
			}
		}
		leaderboardEntries[i] = &model.LeaderboardEntry{
			PlayerID:    participant.PlayerID,
			PlayerName:  playerName,
			Elo:         participant.Elo,
			GamesPlayed: participant.GamesPlayed,
		}
	}

	return leaderboardEntries, nil
}
