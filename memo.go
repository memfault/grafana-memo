package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/benbjohnson/clock"
	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nlopes/slack"
)

var errEmpty = errors.New("empty message")
var errFixedTag = errors.New("cannot override author or chan tag")

type MemoPlugin struct {
	esConn *elastigo.Conn
	re     *regexp.Regexp
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
	fmt.Println("MEMO", ch, usr, out[1])
	memo, err := ParseMemo(ch, usr, out[1], clock.New())
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage("bad memo request: "+err.Error(), msg.Channel))
	}

	response, _ := m.esConn.Index("memos", "memo", "", nil, memo)
	if response.Created {
		rtm.SendMessage(rtm.NewOutgoingMessage("Memo memorized", msg.Channel))
	} else {
		rtm.SendMessage(rtm.NewOutgoingMessage("memo fail :'(", msg.Channel))
	}
}
