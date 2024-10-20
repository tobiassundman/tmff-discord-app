package controller

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
	"tmff-discord-app/internal/app/config"
	"tmff-discord-app/internal/app/repository"
	repomodel "tmff-discord-app/internal/app/repository/model"
	"tmff-discord-app/internal/app/services"
	"tmff-discord-app/internal/app/services/model"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type FanFaction struct {
	gameService        *services.Game
	playerRepo         *repository.Player
	leaderboardService *services.Leaderboard
	gameScraper        *services.GameScraper
	conf               *config.Config
	commandLock        sync.Mutex
	lastCommandByUser  map[string]time.Time
}

func NewFanFaction(
	conf *config.Config,
	playerRepo *repository.Player,
	gameService *services.Game,
	leaderboardService *services.Leaderboard,
	gameScraper *services.GameScraper,
) *FanFaction {
	return &FanFaction{
		gameService:        gameService,
		leaderboardService: leaderboardService,
		gameScraper:        gameScraper,
		conf:               conf,
		playerRepo:         playerRepo,
		lastCommandByUser:  map[string]time.Time{},
	}
}

func (g *FanFaction) FanFactionCommands() (
	[]*discordgo.ApplicationCommand,
	map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate),
) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "register-game",
			Description: "There has to be at least two registered participants in the game.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game-link",
					Description: "The link to the game on Board Game Arena, e.g. https://boardgamearena.com/table?table=571581855",
					Required:    true,
				},
			},
		},
		{
			Name:        "add-player",
			Description: "Add a player.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The username of the player.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The BGA ID of the player.",
					Required:    true,
				},
			},
		},
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"register-game": g.RegisterGame,
		"add-player":    g.AddPlayer,
	}
	return commands, commandHandlers
}

func (g *FanFaction) RegisterGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	g.commandLock.Lock()
	defer g.commandLock.Unlock()
	rateLimitUser := g.rateLimitUser(s, i.Member.User.ID, i.Member.Roles)
	if rateLimitUser {
		err := errors.New("you are limited to one command per hour, ask a moderator to issue the command for you")
		g.respondWithError(s, i, err)
		return
	}
	log.Println("registering game")

	go func() {
		responseMessage, err := g.registerGameAsync(s, i)
		if err != nil {
			g.sendErrorMessage(s, err, i.Member.User.ID)
		} else {
			g.sendAsyncResponse(s, responseMessage)
		}
	}()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@%s> registering game, please wait", i.Member.User.ID),
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %v", err)
	}
}

func (g *FanFaction) AddPlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	g.commandLock.Lock()
	defer g.commandLock.Unlock()
	log.Println("adding player")

	if !g.hasRole(s, i.Member.Roles, "Moderator") {
		err := errors.New("you do not have permission to add a player")
		g.respondWithError(s, i, err)
		return
	}

	rateLimitUser := g.rateLimitUser(s, i.Member.User.ID, i.Member.Roles)
	if rateLimitUser {
		err := errors.New("you are limited to one command per hour, ask a moderator to issue the command for you")
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
			Content: fmt.Sprintf("<@%s> added player %s with ID %s", i.Member.User.ID, playerName, playerID),
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

func (g *FanFaction) sendErrorMessage(s *discordgo.Session, err error, playerID string) {
	log.Printf("could not register game: %v", err)
	gamesChannelID, getChannelErr := getChannelIDByName(s, g.conf.Discord.GuildID, "games")
	if getChannelErr != nil {
		log.Printf("could not get games channel ID: %v", getChannelErr)
	}
	_, sendErr := s.ChannelMessageSend(gamesChannelID, fmt.Sprintf("<@%s> Error: %s", playerID, err.Error()))
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
			Content: fmt.Sprintf("<@%s> %s", i.Member.User.ID, err.Error()),
		},
	})
	if respondErr != nil {
		log.Printf("could not respond to interaction: %v", respondErr)
	}
}

func formatPlayers(players []*repomodel.Player) string {
	var sb strings.Builder
	sb.WriteString("Players\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-20s %-10s %-20s\n", "Name", "ID", "Link"))
	sb.WriteString(fmt.Sprintf("%-20s %-10s %-20s\n", "--------------------", "----------", "--------------------"))
	for _, player := range players {
		sb.WriteString(
			fmt.Sprintf(
				"%-20s %-10s %-20s\n",
				player.Name,
				player.BGAID,
				fmt.Sprintf("https://boardgamearena.com/player?id=%s", player.BGAID),
			),
		)
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
		return "", errors.Wrap(err, "invalid game link")
	}

	gameOutcome, err := g.gameScraper.ExtractGameOutcome(gameLink)
	if err != nil {
		return "", errors.Wrap(err, "could not extract game outcome")
	}

	gameResult, err := g.gameService.RegisterGame(gameOutcome)
	if err != nil {
		return "", errors.Wrap(err, "could not register game")
	}

	err = g.UpdateLeaderboard(s, i.GuildID, "leaderboard")
	if err != nil {
		return "", errors.Wrap(err, "could not update leaderboard")
	}

	return formatGameResult(i, gameResult, gameOutcome.BGALink()), nil
}

func formatGameResult(i *discordgo.InteractionCreate, gameResult []*model.PlayerEloResult, bgaLink string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Thank you for registering a [game](%s) <@%s>!", bgaLink, i.Member.User.ID))
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
		return "", errors.New("game-link option not provided")
	}
	return gameLink.StringValue(), nil
}

func getChannelIDByName(s *discordgo.Session, guildID, channelName string) (string, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", errors.Wrap(err, "error fetching channels")
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("channel with name %s not found", channelName)
}

func (g *FanFaction) UpdateLeaderboard(s *discordgo.Session, guildID, channelName string) error {
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
			return errors.Wrap(err, "could not send message")
		}
	} else {
		_, err = s.ChannelMessageEdit(channelID, messageID, newContent)
		if err != nil {
			return errors.Wrap(err, "could not edit message")
		}
	}

	return nil
}

func getMessageIDContaining(s *discordgo.Session, channelID, searchString string) (string, error) {
	maxMessages := 100
	messages, err := s.ChannelMessages(channelID, maxMessages, "", "", "")
	if err != nil {
		return "", errors.Wrap(err, "could not get messages")
	}

	for _, message := range messages {
		if strings.Contains(message.Content, searchString) {
			return message.ID, nil
		}
	}

	return "", fmt.Errorf("message containing %s not found", searchString)
}

func (g *FanFaction) rateLimitUser(s *discordgo.Session, discordID string, roles []string) bool {
	if g.hasRole(s, roles, "Moderator") {
		return false
	}

	lastCommandTime, ok := g.lastCommandByUser[discordID]
	if ok {
		elapsed := time.Since(lastCommandTime)
		if elapsed < 60*time.Minute {
			return true
		}
	}

	g.lastCommandByUser[discordID] = time.Now()
	return false
}
