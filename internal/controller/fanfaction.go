package controller

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/url"
	"sort"
	"strings"
	"tmff-discord-app/internal/app/config"
	"tmff-discord-app/internal/app/repository"
	repomodel "tmff-discord-app/internal/app/repository/model"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/internal/app/services/model"
)

type FanFaction struct {
	gameService        *services.Game
	playerRepo         *repository.Player
	leaderboardService *services.Leaderboard
	gameScraper        *services.GameScraper
	conf               *config.Config
}

func NewFanFaction(conf *config.Config, playerRepo *repository.Player, gameService *services.Game, leaderboardService *services.Leaderboard, gameScraper *services.GameScraper) *FanFaction {
	return &FanFaction{
		gameService:        gameService,
		leaderboardService: leaderboardService,
		gameScraper:        gameScraper,
		conf:               conf,
		playerRepo:         playerRepo,
	}
}

func (g *FanFaction) RegisterGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("registering game")

	go func() {
		responseMessage, err := g.registerGameAsync(s, i)
		if err != nil {
			g.sendErrorMessage(s, err)
		} else {
			g.sendAsyncResponse(s, responseMessage)
		}
	}()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Game being registered",
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %v", err)
	}
}

func (g *FanFaction) sendErrorMessage(s *discordgo.Session, err error) {
	log.Printf("could not register game: %v", err)
	gamesChannelID, getChannelErr := getChannelIDByName(s, g.conf.Discord.GuildID, "games")
	if getChannelErr != nil {
		log.Printf("could not get games channel ID: %v", getChannelErr)
	}
	_, sendErr := s.ChannelMessageSend(gamesChannelID, "Error: "+err.Error())
	if sendErr != nil {
		log.Printf("could not send game outcome to games channel: %v", sendErr)
	}
}

func (g *FanFaction) sendAsyncResponse(s *discordgo.Session, message string) {
	gamesChannelID, getChannelErr := getChannelIDByName(s, g.conf.Discord.GuildID, "games")
	if getChannelErr != nil {
		log.Printf("could not get games channel ID: %v", getChannelErr)
	}
	_, sendErr := s.ChannelMessageSend(gamesChannelID, message)
	if sendErr != nil {
		log.Printf("could not send game outcome to games channel: %v", sendErr)
	}
}

func (g *FanFaction) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: err.Error(),
		},
	})
	if respondErr != nil {
		log.Printf("could not respond to interaction: %v", respondErr)
	}
}

func (g *FanFaction) AddPlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("adding player")

	if !g.hasRole(s, i.Member.Roles, "Moderator") {
		err := fmt.Errorf("you do not have permission to add a player")
		g.respondWithError(s, i, err)
		return
	}

	playerName, err := g.getOption(i, "name")
	if err != nil {
		g.respondWithError(s, i, err)
		return
	}
	playerID, err := g.getOption(i, "id")
	if err != nil {
		g.respondWithError(s, i, err)
		return
	}

	err = g.playerRepo.InsertPlayer(playerName, playerID)
	if err != nil {
		g.respondWithError(s, i, err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Player added",
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %v", err)
	}

	players, err := g.playerRepo.GetPlayers()
	if err != nil {
		log.Printf("could not get players: %v", err)
		return
	}

	sort.Slice(players, func(i, j int) bool {
		return strings.ToLower(players[i].Name) < strings.ToLower(players[j].Name)
	})

	registeredPlayersChannelID, getChannelErr := getChannelIDByName(s, i.GuildID, "registered-players")
	if getChannelErr != nil {
		log.Printf("could not get registered players channel ID: %v", getChannelErr)
		return
	}

	err = upsertMessage(s, registeredPlayersChannelID, "Players", formatPlayers(players))
	if err != nil {
		log.Printf("could not update players: %v", err)
	}
}

func formatPlayers(players []*repomodel.Player) string {
	var sb strings.Builder
	sb.WriteString("Players\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-20s %-10s %-20s\n", "Name", "ID", "Link"))
	sb.WriteString(fmt.Sprintf("%-20s %-10s %-20s\n", "--------------------", "----------", "--------------------"))
	for _, player := range players {
		sb.WriteString(fmt.Sprintf("%-20s %-10s %-20s\n", player.Name, player.BGAID, fmt.Sprintf("https://boardgamearena.com/player?id=%s", player.BGAID)))
	}
	sb.WriteString("```\n")
	return sb.String()
}

func (g *FanFaction) hasRole(s *discordgo.Session, roleIDs []string, requiredRoleName string) bool {
	for _, roleID := range roleIDs {
		role, err := s.State.Role(g.conf.Discord.GuildID, roleID)
		if err != nil {
			log.Printf("could not get role: %v", err)
			continue
		}
		if role.Name == requiredRoleName {
			return true
		}
	}
	return false
}

func (g *FanFaction) registerGameAsync(s *discordgo.Session, i *discordgo.InteractionCreate) (string, error) {
	gameLink, err := g.getOption(i, "game-link")
	if err != nil {
		return "", err
	}
	_, err = url.Parse(gameLink)
	if err != nil {
		return "", fmt.Errorf("invalid game link: %v", err)
	}

	gameOutcome, err := g.gameScraper.ExtractGameOutcome(gameLink)
	if err != nil {
		return "", fmt.Errorf("could not extract game outcome: %v", err)
	}

	gameResult, err := g.gameService.RegisterGame(gameOutcome)
	if err != nil {
		return "", fmt.Errorf("could not register game: %v", err)
	}

	err = g.updateLeaderboard(s, i.GuildID, "leaderboard")
	if err != nil {
		return "", fmt.Errorf("could not update leaderboard: %v", err)
	}

	return formatGameResult(gameResult), nil
}

func formatGameResult(gameResult []*model.PlayerEloResult) string {
	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-5s %-20s %-10s\n", "Rank", "Name", "Elo Change"))
	sb.WriteString(fmt.Sprintf("%-5s %-20s %-10s\n", "-----", "--------------------", "----------"))
	for i, result := range gameResult {
		sb.WriteString(fmt.Sprintf("%-5d %-20s %-10d\n", i+1, result.Name, result.EloChange))
	}
	sb.WriteString("```\n")
	return sb.String()
}

func (g *FanFaction) getOption(i *discordgo.InteractionCreate, optionName string) (string, error) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	gameLink, ok := optionMap[optionName]
	if !ok {
		return "", fmt.Errorf("game-link option not provided")
	}
	return gameLink.StringValue(), nil
}

func getChannelIDByName(s *discordgo.Session, guildID, channelName string) (string, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", fmt.Errorf("error fetching channels: %v", err)
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("channel with name %s not found", channelName)
}

func (g *FanFaction) updateLeaderboard(s *discordgo.Session, guildID, channelName string) error {
	leaderboard, err := g.leaderboardService.GetLeaderboard()
	if err != nil {
		return err
	}

	leaderboardChannelID, getChannelErr := getChannelIDByName(s, guildID, channelName)
	if getChannelErr != nil {
		return getChannelErr
	}

	leaderboardUpdate := leaderboard.String()
	log.Printf("Leaderboard: %s", leaderboardUpdate)

	err = upsertMessage(s, leaderboardChannelID, "Leaderboard", fmt.Sprintf("```\n%s\n```", leaderboardUpdate))
	if err != nil {
		return err
	}
	return nil
}

func upsertMessage(s *discordgo.Session, channelID, messageHeader, newContent string) error {
	messageID, err := getMessageIDContaining(s, channelID, messageHeader)
	if err != nil {
		_, err = s.ChannelMessageSend(channelID, newContent)
		if err != nil {
			return fmt.Errorf("could not send new message: %v", err)
		}
	} else {
		_, err = s.ChannelMessageEdit(channelID, messageID, newContent)
		if err != nil {
			return fmt.Errorf("could not update message: %v", err)
		}
	}

	return nil
}

func getMessageIDContaining(s *discordgo.Session, channelID, searchString string) (string, error) {
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return "", fmt.Errorf("error fetching messages: %v", err)
	}

	for _, message := range messages {
		if strings.Contains(message.Content, searchString) {
			return message.ID, nil
		}
	}

	return "", fmt.Errorf("message containing %s not found", searchString)
}
