package daemon

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/nlopes/slack"
	"github.com/raintank/memo/store"
	log "github.com/sirupsen/logrus"
)

var errEmpty = errors.New("empty message")
var errFixedTag = errors.New("cannot override author or chan tag")
var helpMessage = "Hi. I only support memo requests. See https://github.com/raintank/memo/blob/master/README.md#message-format"

type Daemon struct {
	apiToken string
	re       *regexp.Regexp
	api      *slack.Client
	info     *slack.Info
	rtm      *slack.RTM
	store    store.Store
}

func New(apiToken string, store store.Store) *Daemon {
	d := Daemon{
		apiToken: apiToken,
		re:       regexp.MustCompile("^memo (.*)"),
		store:    store,
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
	memo, err := ParseMemo(ch, usr, out[1], clock.New())
	if err != nil {
		log.Infof("Received invalid memo request on channel %s, from user %s. message is: %s", ch, usr, out[1])
		d.rtm.SendMessage(d.rtm.NewOutgoingMessage("bad memo request: "+err.Error(), msg.Channel))
		return
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
	for _, ch := range d.info.Channels {
		if ch.ID == id {
			return ch.Name
		}
	}
	return "null"
}

func (d *Daemon) userIdToName(id string) string {
	for _, usr := range d.info.Users {
		if usr.ID == id {
			return usr.Name
		}
	}
	return "null"
}
