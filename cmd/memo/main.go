package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/raintank/memo/cfg"
	"github.com/raintank/memo/daemon"
	"github.com/raintank/memo/store"
	log "github.com/sirupsen/logrus"
)

var configFile = "/etc/memo.toml"

func main() {
	if len(os.Args) > 2 {
		log.Fatal("usage: memo [path-to-config]")
	}
	if len(os.Args) == 2 {
		configFile = os.Args[1]
	}

	var config cfg.Config
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		log.Fatalf("Invalid config file %q: %s", configFile, err.Error())
	}

	lvl, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to parse log-level %q: %s", config.LogLevel, err.Error())
	}
	log.SetLevel(lvl)
	log.SetOutput(os.Stdout)

	store, err := store.NewGrafana(config.Grafana.ApiKey, config.Grafana.ApiUrl)
	if err != nil {
		log.Fatalf("failed to create Grafana store: %s", err.Error())
	}
	daemon := daemon.New(config.Slack.ApiToken, store)

	daemon.Run()
}
