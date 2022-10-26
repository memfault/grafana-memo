package daemon

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/grafana/memo/cfg"
	"github.com/grafana/memo/parser"
	discordService "github.com/grafana/memo/service/discord"
	slackService "github.com/grafana/memo/service/slack"
	"github.com/grafana/memo/store"
	log "github.com/sirupsen/logrus"
)

// Daemon
type Daemon struct {
	// store
	store store.Store
	// config
	config cfg.Config
	// parser
	parser parser.Parser
}

// New
func New(config cfg.Config, store store.Store) *Daemon {
	d := Daemon{
		store:  store,
		config: config,
		parser: parser.New(),
	}

	return &d
}

// Run
func (d *Daemon) Run() {
	log.Info("Memo starting")

	if d.config.Slack.Enabled {
		log.Info("slack enabled")
		_, err := slackService.New(
			d.config.Slack,
			d.parser,
			d.store,
		)

		if err != nil {
			log.Fatalf("Could not initialise Slack Handler")
		}
	}

	if d.config.Discord.Enabled {
		log.Info("discord enabled")
		_, err := discordService.New(
			d.config.Discord,
			d.parser,
			d.store,
		)

		if err != nil {
			log.Fatalf("could not initialise discord handler")
		}
	}

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	// hold the process open until we panic or cancel
	<-gracefulStop

	log.Info("shutting down")
	os.Exit(0)
}
