package main

import (
	"errors"
	"fmt"
	"github.com/benbjohnson/clock"
	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/nlopes/slack"
	"github.com/raintank/dur"
	"regexp"
	"sort"
	"strings"
	"time"
)

var errEmpty = errors.New("empty message")
var errFixedTag = errors.New("cannot override author or chan tag")

type MemoPlugin struct {
	esConn *elastigo.Conn
	re     *regexp.Regexp
}

type Memo struct {
	Date time.Time
	Desc string
	Tags []string
}

func getMemo(ch, usr, msg string, cl clock.Clock) (*Memo, error) {
	msg = strings.TrimSpace(msg)
	words := strings.Fields(msg)
	if len(words) == 0 {
		return nil, errEmpty
	}

	ts := cl.Now().Add(-25 * time.Second)
	dur, err := dur.ParseDuration(words[0])
	if err == nil {
		ts = cl.Now().Add(-time.Duration(dur) * time.Second)
		msg = msg[len(words[0]):]
	} else {
		parsed, err := time.Parse(time.RFC3339, words[0])
		if err == nil {
			ts = parsed
			msg = msg[len(words[0]):]
		}
	}
	extraTags := make([]string, 0)
	stripChars := 0
	for i := len(words) - 1; i >= 0 && strings.Contains(words[i], ":"); i-- {
		stripChars += len(words[i] + " ") // this is a bug, should account for true amount of whitespace. sue me
		cleanTag := strings.TrimSpace(words[i])
		if strings.HasPrefix(cleanTag, "author:") || strings.HasPrefix(cleanTag, "chan:") {
			return nil, errFixedTag
		}
		extraTags = append(extraTags, strings.TrimSpace(words[i]))
	}
	msg = msg[:len(msg)-stripChars]
	msg = strings.TrimSpace(msg)
	if len(msg) == 0 {
		return nil, errEmpty
	}
	tags := []string{
		"author:" + usr,
	}
	if ch != "null" {
		tags = append(tags, "chan:"+ch)
	}
	sort.Strings(extraTags)
	tags = append(tags, extraTags...)

	return &Memo{
		ts.UTC(),
		msg,
		tags,
	}, nil
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
	memo, err := getMemo(ch, usr, out[1], clock.New())
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
