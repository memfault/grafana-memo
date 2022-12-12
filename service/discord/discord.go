package discord

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	mem "github.com/grafana/memo"
	"github.com/grafana/memo/cfg"
	"github.com/grafana/memo/parser"
	"github.com/grafana/memo/service"
	"github.com/grafana/memo/store"
)

// DiscordService
type DiscordService struct {
	// config
	config cfg.Discord

	// parser takes the memo and extracts the values from it
	parser parser.Parser
	// store puts the memo in the defined store
	store store.Store

	// client for communicating with discord API
	client *discordgo.Session
}

// Name returns the basic name of this service
func (d DiscordService) Name() string {
	return "discord"
}

// handleMessage takes the discord message event and creates the memo, to pass
// to the store for storing the memo
func (d *DiscordService) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	log.Debugf("new discord message: %v", m.Content)
	memo, err := d.parser.Parse(m.Content)
	if err != nil {
		if err.Error() != mem.ErrEmpty.Error() {
			d.client.ChannelMessageSend(m.ChannelID, fmt.Sprintf("memo failed: %s", err.Error()))
		}

		return
	}

	if memo == nil {
		return
	}

	tags := []string{
		"author:" + m.Author.Username,
		"chan:" + m.ChannelID,
		"source:discord",
	}

	memo.BuildTags(tags)

	err = d.store.Save(*memo)
	if err != nil {
		d.client.ChannelMessageSend(m.ChannelID, fmt.Sprintf("memo failed: %s", err.Error()))
		return
	}

	d.client.ChannelMessageSend(m.ChannelID, "Memo saved!")
}

// New creates a new instance of this service
func New(config cfg.Discord, parser parser.Parser, store store.Store) (service.Service, error) {
	client, err := discordgo.New("Bot " + config.BotToken)
	if err != nil {
		log.Fatalf("error connecting to discord: %s", err.Error())
	}

	d := DiscordService{
		config: config,
		parser: parser,
		store:  store,
		client: client,
	}

	d.client.AddHandler(d.handleMessage)
	d.client.Identify.Intents |= discordgo.IntentsGuildMessages
	d.client.Identify.Intents |= discordgo.IntentMessageContent

	go func() {
		err := d.client.Open()
		if err != nil {
			log.Fatalf("discord connection failed: %s", err.Error())
		}
	}()

	return d, nil
}
