package client

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"log"
	"tmff-discord-app/internal/app/config"
)

type Discord struct {
	client *discordgo.Session
	conf   *config.Config
}

func NewDiscord(conf *config.Config) (*Discord, error) {
	client, err := discordgo.New("Bot " + conf.Discord.Token)
	if err != nil {
		return nil, errors.Wrap(err, "could not create discord client")
	}
	return &Discord{
		client: client,
		conf:   conf,
	}, nil
}

func (d *Discord) Initialize(
	commands []*discordgo.ApplicationCommand,
	commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate),
) error {
	err := d.client.Open()
	if err != nil {
		return errors.Wrap(err, "could not open discord connection")
	}

	d.client.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	d.client.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})

	for _, command := range commands {
		_, err = d.client.ApplicationCommandCreate(d.conf.Discord.AppID, d.conf.Discord.GuildID, command)
		if err != nil {
			return errors.Wrap(err, "could not create application command")
		}
	}

	return nil
}

func (d *Discord) Close() error {
	return d.client.Close()
}
