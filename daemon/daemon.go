package daemon

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/grafana/memo"
	"github.com/grafana/memo/store"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

var errEmpty = errors.New("empty message")
var helpMessage = "Hi. I only support memo requests. See https://github.com/grafana/memo/blob/master/README.md#message-format"

type Daemon struct {
	apiToken string
	re       *regexp.Regexp
	api      *slack.Client
	info     *slack.Info
	rtm      *slack.RTM
	store    store.Store

	// see https://github.com/nlopes/slack/issues/532
	chanIdToNameCache map[string]string
	userIdToNameCache map[string]string
}

func New(apiToken string, store store.Store) *Daemon {
	d := Daemon{
		apiToken: apiToken,
		re:       regexp.MustCompile("^memo (.*)"),
		store:    store,

		chanIdToNameCache: make(map[string]string),
		userIdToNameCache: make(map[string]string),
	}
	return &d
}

func (d *Daemon) Run() {
	log.Info("Memo starting")
	d.api = slack.New(d.apiToken)
	d.rtm = d.api.NewRTM()
	go d.rtm.ManageConnection()
	for {
		select {
		case msg := <-d.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectingEvent:
				log.Infof("Connecting to slack. attempt %d", ev.Attempt)

			case *slack.ConnectedEvent:
				log.Info("Connected to slack")
				d.info = ev.Info

			case *slack.HelloEvent:
				log.Info("Received hello from slack")

			case *slack.MessageEvent:
				d.handleMessage(ev.Msg)

			case *slack.PresenceChangeEvent:
				log.Infof("Presence changed to type:%s - presence:%s - user%s", ev.Type, ev.Presence, ev.User)

			case *slack.LatencyReport:
				log.Debugf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				log.Warnf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Fatalf("Invalid credentials")

			default:
				// lots of other kinds of messages that we can ignore
			}
		}
	}
}

func (d *Daemon) handleMessage(msg slack.Msg) {
	ch := d.chanIdToName(msg.Channel)
	usr := d.userIdToName(msg.User)

	out := d.re.FindStringSubmatch(msg.Text)
	if len(out) == 0 {
		if strings.HasPrefix(msg.Text, "memo:") || strings.HasPrefix(msg.Text, "mrbot:") {
			d.rtm.SendMessage(d.rtm.NewOutgoingMessage(helpMessage, msg.Channel))
			return
		}
		// we're in a private message. anything the user says is for us
		if ch == "null" {
			d.rtm.SendMessage(d.rtm.NewOutgoingMessage(helpMessage, msg.Channel))
			return
		}
		// we're in a channel. don't spam in it. the message was probably not meant for us.
		return
	}
	tags := []string{
		"memo",
		"author:" + usr,
		"chan:" + ch,
	}

	ts, desc, extraTags, err := ParseCommand(out[1], clock.New())
	if err != nil {
		log.Infof("Received invalid memo request on channel %s, from user %s. message is: %s", ch, usr, out[1])
		d.rtm.SendMessage(d.rtm.NewOutgoingMessage("bad memo request: "+err.Error(), msg.Channel))
		return
	}
	tags, err = memo.BuildTags(tags, extraTags)
	if err != nil {
		log.Infof("Received invalid memo request on channel %s, from user %s. message is: %s", ch, usr, out[1])
		d.rtm.SendMessage(d.rtm.NewOutgoingMessage("bad memo request: "+err.Error(), msg.Channel))
		return
	}
	memo := memo.Memo{
		Date: ts.UTC(),
		Desc: desc,
		Tags: tags,
	}
	log.Infof("Received a valid memo request on channel %s, from user %s. message is: %s", ch, usr, out[1])

	err = d.store.Save(memo)
	if err != nil {
		d.rtm.SendMessage(d.rtm.NewOutgoingMessage(fmt.Sprintf("memo failed: %s", err.Error()), msg.Channel))
		return
	}

	d.rtm.SendMessage(d.rtm.NewOutgoingMessage("Memo saved", msg.Channel))
}

func (d *Daemon) chanIdToName(id string) string {
	name, ok := d.chanIdToNameCache[id]
	if ok {
		return name
	}
	g, err := d.api.GetChannelInfo(id)
	if err != nil {
		return "null"
	}
	d.chanIdToNameCache[id] = g.Name
	return g.Name
}

func (d *Daemon) userIdToName(id string) string {
	name, ok := d.userIdToNameCache[id]
	if ok {
		return name
	}
	u, err := d.api.GetUserInfo(id)
	if err != nil {
		return "null"
	}
	d.userIdToNameCache[id] = u.Name
	return u.Name
}
