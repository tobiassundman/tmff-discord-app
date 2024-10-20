package client

import (
	"log"
	"time"
	"tmff-discord-app/internal/app/config"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type Discord struct {
	Client *discordgo.Session
	conf   *config.Config
}

func NewDiscord(conf *config.Config) (*Discord, error) {
	client, err := discordgo.New("Bot " + conf.Discord.Token)
	if err != nil {
		return nil, errors.Wrap(err, "could not create discord Client")
	}
	return &Discord{
		Client: client,
		conf:   conf,
	}, nil
}

func (d *Discord) Initialize(
	commands []*discordgo.ApplicationCommand,
	commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate),
) error {
	err := d.Client.Open()
	if err != nil {
		return errors.Wrap(err, "could not open discord connection")
	}

	d.Client.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	d.Client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})

	for _, command := range commands {
		_, err = d.Client.ApplicationCommandCreate(d.conf.Discord.AppID, d.conf.Discord.GuildID, command)
		if err != nil {
			return errors.Wrap(err, "could not create application command")
		}
	}

	return nil
}

func (d *Discord) UpdateLeaderboard() (string, error) {
	return "", nil
}

func (d *Discord) Close() error {
	return d.Client.Close()
}

func (d *Discord) SetBotStatus() {
	now := time.Now()
	botRunningSince := "Bot running since: " + now.Format("2006-01-02 15:04:05")
	err := d.Client.UpdateCustomStatus(botRunningSince)
	log.Printf("Error setting bot status: %v", err)
}
