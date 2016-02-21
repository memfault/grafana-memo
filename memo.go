package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	elastigo "github.com/mattbaird/elastigo/lib"
	"log"
	"os"
	"strings"
	"time"
)

var esConn *elastigo.Conn

type Memo struct {
	Date time.Time
	Desc string
	Tags []string
}

func main() {
	esConn = elastigo.NewConn()
	esConn.Domain = "localhost"
	esConn.Port = "9200"
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready. my name is", id)

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}
		spew.Dump(m)
		//"2m", "40s", "10min40s"

		if m.Type == "message" {
			if strings.HasPrefix(m.Text, "memo ") && len(m.Text) > 5 {
				//fields := strings.Fields(m.Text)
				go func(m Message) {
					memo := Memo{
						time.Now().UTC(),
						m.Text[5:],
						[]string{"memo"},
					}
					response, _ := esConn.Index("memos", "memo", "", nil, memo)
					if response.Created {
						m.Text = "Memo memorized"
					} else {
						m.Text = "memo fail :'("
					}
					postMessage(ws, m)
				}(m)
			}
		}
	}
}
