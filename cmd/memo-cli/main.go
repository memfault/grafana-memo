package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/grafana/memo"
	"github.com/grafana/memo/cfg"
	"github.com/grafana/memo/store"
)

var configFile string
var timestamp int
var extraTags CsvStringVar
var message string

func main() {

	flag.IntVar(&timestamp, "ts", int(time.Now().Unix()), "unix timestamp. always defaults to 'now'")
	flag.Var(&extraTags, "tags", "One or more comma-separated tags to submit, in addition to 'memo', 'user:<unix-username>' and 'host:<hostname>'")
	flag.StringVar(&message, "msg", "", "message to submit")
	flag.StringVar(&configFile, "config", "~/.memo.toml", "config file location")
	flag.Parse()

	if message == "" {
		fmt.Fprintln(os.Stderr, "message cannot be empty")
		os.Exit(2)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current user: %s\n", err.Error())
		os.Exit(2)
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current hostname: %s\n", err.Error())
		os.Exit(2)
	}

	if strings.HasPrefix(configFile, "~/") {
		configFile = filepath.Join(usr.HomeDir, configFile[2:])
	}

	var config cfg.Config
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config file %q: %s\n", configFile, err.Error())
		os.Exit(2)
	}

	store, err := store.NewGrafana(config.Grafana.ApiKey, config.Grafana.ApiUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create Grafana store: %s\n", err.Error())
		os.Exit(2)
	}
	tags := []string{
		"memo",
		"user:" + usr.Username,
		"host:" + hostname,
	}
	tags, err = memo.BuildTags(tags, extraTags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set tags: %s\n", err.Error())
		os.Exit(2)
	}
	m := memo.Memo{
		Date: time.Unix(int64(timestamp), 0),
		Desc: message,
		Tags: tags,
	}
	err = store.Save(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save memo in store: %s\n", err.Error())
		os.Exit(2)
	}
	fmt.Println("memo saved")
}
