package services

import (
	"github.com/pkg/errors"
	"math"
	"sort"
	"tmff-discord-app/internal/app/repository"
	repomodel "tmff-discord-app/internal/app/repository/model"
	"tmff-discord-app/internal/app/services/model"
)

type PlayerNameToID map[string]string
type PlayerIDToName map[string]string
type PlayerIDToCurrentElo map[string]int
type PlayerIDToEloChange map[string]int
type PlayerIDToScore map[string]int

type Game struct {
	playerRepo *repository.Player
	gameRepo   *repository.Game
	seasonRepo *repository.Season
	kValue     float64
}

func NewGame(playerRepo *repository.Player, gameRepo *repository.Game, seasonRepo *repository.Season, kValue int) *Game {
	return &Game{
		playerRepo: playerRepo,
		gameRepo:   gameRepo,
		seasonRepo: seasonRepo,
		kValue:     float64(kValue),
	}
}

func (g *Game) RegisterGame(gameOutcome *model.GameOutcome) ([]*model.PlayerEloResult, error) {
	_, err := g.gameRepo.GetGameWithParticipants(gameOutcome.ID)
	if !errors.Is(err, repository.ErrGameNotFound) {
		return nil, errors.Wrap(err, "game is already registered")
	}

	registeredPlayers, err := g.getRegisteredPlayers(gameOutcome)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get registered players")
	}
	if len(registeredPlayers) < 2 {
		return nil, errors.New("less than two registered players found for game")
	}

	participantsRating, err := g.getPlayerElos(gameOutcome, registeredPlayers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get player elos")
	}

	eloChangeMap := g.getEloChangeForPlayers(gameOutcome, participantsRating, registeredPlayers)

	for playerID, eloChange := range eloChangeMap {
		_, updateErr := g.seasonRepo.UpsertSeasonParticipant(playerID, eloChange)
		if updateErr != nil {
			return nil, errors.Wrap(updateErr, "failed to update season participant")
		}
	}

	playerScores := playerScoreByID(gameOutcome, registeredPlayers)
	playerNamesByID := playerNameByID(registeredPlayers)

	var gameParticipants []*repomodel.GameParticipant
	var playerEloResults []*model.PlayerEloResult
	for playerID, eloChange := range eloChangeMap {
		gameParticipants = append(gameParticipants, &repomodel.GameParticipant{
			GameID:    gameOutcome.ID,
			PlayerID:  playerID,
			Score:     playerScores[playerID],
			EloChange: eloChange,
			EloBefore: participantsRating[playerID],
		})
		playerEloResults = append(playerEloResults, &model.PlayerEloResult{
			Name:      playerNamesByID[playerID],
			ID:        playerID,
			Score:     playerScores[playerID],
			EloBefore: participantsRating[playerID],
			EloChange: eloChange,
		})
	}
	err = g.gameRepo.CreateGameWithParticipants(gameOutcome.ID, gameParticipants)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create game with participants")
	}
	// Sort the slice by Age using sort.Slice and a custom comparison function
	sort.Slice(playerEloResults, func(i, j int) bool {
		return playerEloResults[i].EloChange > playerEloResults[j].EloChange
	})

	return playerEloResults, nil
}

func playerScoreByID(gameOutcome *model.GameOutcome, idMap PlayerNameToID) PlayerIDToScore {
	playerScore := make(map[string]int)
	for _, player := range gameOutcome.Players {
		playerScore[idMap[player.Name]] = player.Score
	}
	return playerScore
}

func playerNameByID(idMap PlayerNameToID) PlayerIDToName {
	playerName := make(map[string]string)
	for name, id := range idMap {
		playerName[id] = name
	}
	return playerName
}

func (g *Game) getEloChangeForPlayers(gameOutcome *model.GameOutcome, participantsRating PlayerIDToCurrentElo, idMap PlayerNameToID) PlayerIDToEloChange {
	eloChangeForPlayerIDs := make(map[string]int)
	for _, mainPlayer := range gameOutcome.Players {
		if _, ok := participantsRating[idMap[mainPlayer.Name]]; !ok {
			continue
		}
		for _, opponent := range gameOutcome.Players {
			if mainPlayer.Name == opponent.Name {
				continue
			}
			if _, ok := participantsRating[idMap[opponent.Name]]; !ok {
				continue
			}
			var score float64
			switch {
			case mainPlayer.Score > opponent.Score:
				score = 1
			case mainPlayer.Score < opponent.Score:
				score = 0
			case mainPlayer.Score == opponent.Score:
				score = 0.5
			}
			eloChange := g.calculateSubMatchEloChange(participantsRating[idMap[mainPlayer.Name]], participantsRating[idMap[opponent.Name]], score)
			eloChangeForPlayerIDs[idMap[mainPlayer.Name]] += eloChange
		}
	}
	return eloChangeForPlayerIDs
}

func (g *Game) getRegisteredPlayers(gameOutcome *model.GameOutcome) (PlayerNameToID, error) {
	registeredPlayers := make(PlayerNameToID)
	for _, player := range gameOutcome.Players {
		registeredPlayer, getPlayerErr := g.playerRepo.GetPlayer(player.Name)
		if errors.Is(getPlayerErr, repository.ErrPlayerNotFound) {
			continue
		}
		if getPlayerErr != nil {
			return nil, getPlayerErr
		}
		registeredPlayers[player.Name] = registeredPlayer.BGAID
	}
	return registeredPlayers, nil
}

func (g *Game) getPlayerElos(
	gameOutcome *model.GameOutcome,
	registeredPlayers PlayerNameToID,
) (PlayerIDToCurrentElo, error) {
	seasonParticipants, err := g.seasonRepo.GetAll()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get season participants")
	}

	participantsRating := make(map[string]int)
	for _, gameParticipant := range gameOutcome.Players {
		_, ok := registeredPlayers[gameParticipant.Name]
		if !ok {
			continue
		}
		participantsRating[registeredPlayers[gameParticipant.Name]] = repomodel.StartElo
	}
	for _, gameParticipant := range gameOutcome.Players {
		for _, participant := range seasonParticipants {
			playerID, ok := registeredPlayers[gameParticipant.Name]
			if !ok {
				continue
			}
			if playerID == participant.PlayerID {
				participantsRating[playerID] = participant.Elo
			}
		}
	}
	return participantsRating, nil
}

func (g *Game) calculateSubMatchEloChange(playerRating, opponentRating int, actualScore float64) int {
	subMatchCount := 3.0
	const K = 64
	expectedScore := 1 / (1 + math.Pow(10, float64(opponentRating-playerRating)/400))
	eloChange := (g.kValue * (actualScore - expectedScore)) / subMatchCount
	return int(math.Round(eloChange))
}
