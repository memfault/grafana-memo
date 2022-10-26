package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/grafana/memo/cfg"
	"github.com/grafana/memo/daemon"
	"github.com/grafana/memo/store"
	log "github.com/sirupsen/logrus"
)

var configFile = "/etc/memo.toml"

// main
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

	store, err := store.NewGrafana(config.Grafana.ApiKey, config.Grafana.ApiUrl, config.Grafana.TLSKey, config.Grafana.TLSCert)
	if err != nil {
		log.Fatalf("failed to create Grafana store: %s", err.Error())
	}
	err = store.Check()
	if err != nil {
		log.Fatalf("Grafana store is unhealthy: %s", err.Error())
	}

	daemon := daemon.New(config, store)

	daemon.Run()
}
