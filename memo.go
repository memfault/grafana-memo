package main

import (
	"fmt"
	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nlopes/slack"
	"regexp"
	"time"
)

type MemoPlugin struct {
	esConn *elastigo.Conn
	re     *regexp.Regexp
}

type Memo struct {
	Date time.Time
	Desc string
	Tags []string
}

func (m *MemoPlugin) Start() {
	m.re = regexp.MustCompile("^memo (.*)")
	m.esConn = elastigo.NewConn()
	m.esConn.Domain = "localhost"
	m.esConn.Port = "9200"
}

func (m *MemoPlugin) Handle(msg slack.Msg) {
	out := m.re.FindStringSubmatch(msg.Text)
	if len(out) == 0 {
		return
	}
	ch := chanIdToName(msg.Channel)
	usr := userIdToName(msg.User)

	// "category" == channel, but if it was a PM, just use the username
	category := usr
	if ch != "null" {
		category = ch
	}
	fmt.Println("MEMO", ch, usr, out[1])

	//"2m", "40s", "10min40s"

	memo := Memo{
		time.Now().UTC(),
		out[1],
		[]string{category, "author:" + usr},
	}
	response, _ := m.esConn.Index("memos", "memo", "", nil, memo)
	if response.Created {
		rtm.SendMessage(rtm.NewOutgoingMessage("Memo memorized", msg.Channel))
	} else {
		rtm.SendMessage(rtm.NewOutgoingMessage("memo fail :'(", msg.Channel))
	}
}
