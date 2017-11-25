package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"os"
)

var api *slack.Client
var info *slack.Info
var rtm *slack.RTM

func chanIdToName(id string) string {
	for _, ch := range info.Channels {
		if ch.ID == id {
			return ch.Name
		}
	}
	return "null"
}

func userIdToName(id string) string {
	for _, usr := range info.Users {
		if usr.ID == id {
			return usr.Name
		}
	}
	return "null"
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}
	memo := &MemoPlugin{}
	memo.Start()
	api = slack.New(os.Args[1])
	//api.SetDebug(true)
	rtm = api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// Ignore hello

			case *slack.ConnectedEvent:
				info = ev.Info

			case *slack.MessageEvent:
				memo.Handle(ev.Msg)

			case *slack.PresenceChangeEvent:
			//	fmt.Printf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
			//	fmt.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Print("Invalid credentials")
				break Loop

			default:
				//	fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}
