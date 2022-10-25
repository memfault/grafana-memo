package daemon

import (
	"errors"
	"fmt"
	llog "log"
	"os"
	"regexp"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/grafana/memo"
	"github.com/grafana/memo/store"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var ErrEmpty = errors.New("empty message")
var helpMessage = "Hi. I only support memo requests. See https://github.com/grafana/memo/blob/master/README.md#message-format"

type Daemon struct {
	botToken string
	appToken string

	re     *regexp.Regexp
	api    *slack.Client
	info   *slack.Info
	socket *socketmode.Client
	store  store.Store

	// see https://github.com/nlopes/slack/issues/532
	chanIdToNameCache map[string]string
	userIdToNameCache map[string]string
}

func New(botToken string, appToken string, store store.Store) *Daemon {
	d := Daemon{
		botToken: botToken,
		appToken: appToken,

		re:    regexp.MustCompile("^memo (.*)"),
		store: store,

		chanIdToNameCache: make(map[string]string),
		userIdToNameCache: make(map[string]string),
	}
	return &d
}

func (d *Daemon) Run() {
	log.Info("Memo starting")
	d.api = slack.New(
		d.botToken,
		slack.OptionAppLevelToken(d.appToken),
	)

	d.socket = socketmode.New(
		d.api,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(llog.New(os.Stdout, "slack", llog.Lshortfile|llog.LstdFlags)),
	)

	go func() {
		for evt := range d.socket.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Info("Connecting to slack")
			case socketmode.EventTypeConnectionError:
				log.Errorf("Connection error: %v", evt)
			case socketmode.EventTypeConnected:
				log.Info("Socket connected")
			case socketmode.EventTypeDisconnect:
				log.Info("Socket disconnected")
			case socketmode.EventTypeIncomingError:
				log.Errorf("Connection error: %v", evt)
			case socketmode.EventTypeHello:
				log.Info("Received hello from slack, hi!")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				d.socket.Ack(*evt.Request)
				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						d.handleMessage(ev)
					}
				}
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	log.Printf("Running")
	err := d.socket.Run()
	log.Errorf("Stopping: %s", err.Error())
}

func (d *Daemon) handleMessage(msg *slackevents.MessageEvent) {
	ch := d.chanIdToName(msg.Channel)
	usr := d.userIdToName(msg.User)

	out := d.re.FindStringSubmatch(msg.Text)
	if len(out) == 0 {
		// handle the case of a user typing <our-username>: some message
		if strings.HasPrefix(msg.Text, "memo:") || strings.HasPrefix(msg.Text, "mrbot:") || strings.HasPrefix(msg.Text, "memobot:") {
			log.Debugf("A user seems to direct a message %q to us, but we don't understand it. so sending help message back", msg.Text)
			d.api.PostMessage(msg.Channel, slack.MsgOptionPostEphemeral(msg.User), slack.MsgOptionText(helpMessage, false))
			return
		}
		// we're in a private message. anything the user says is for us
		if ch == "null" {
			log.Debugf("A user sent us a DM %q but we don't understand it. so sending help message back", msg.Text)
			d.api.PostMessage(msg.Channel, slack.MsgOptionPostEphemeral(msg.User), slack.MsgOptionText(helpMessage, false))
			return
		}
		// we're in a channel. don't spam in it. the message was probably not meant for us.
		log.Tracef("Received message %q, not for us. ignoring", msg.Text)
		return
	}

	tags := []string{
		"memo",
		"author:" + usr,
		"chan:" + ch,
	}

	log.Debugf("A user sent us the command %q", msg.Text)

	ts, desc, extraTags, err := ParseCommand(out[1], clock.New())
	if err != nil {
		log.Infof("Received invalid memo request on channel %s, from user %s. message is: %s", ch, usr, out[1])
		d.api.PostMessage(msg.Channel, slack.MsgOptionPostEphemeral(msg.User), slack.MsgOptionText("bad memo request: "+err.Error(), false))
		return
	}
	tags, err = memo.BuildTags(tags, extraTags)
	if err != nil {
		log.Infof("Received invalid memo request on channel %s, from user %s. message is: %s", ch, usr, out[1])
		d.api.PostMessage(msg.Channel, slack.MsgOptionPostEphemeral(msg.User), slack.MsgOptionText("bad memo request: "+err.Error(), false))
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
		d.api.PostMessage(msg.Channel, slack.MsgOptionPostEphemeral(msg.User), slack.MsgOptionText("memo failed: "+err.Error(), false))
		return
	}

	d.api.PostMessage(msg.Channel, slack.MsgOptionPostEphemeral(msg.User), slack.MsgOptionText("Memo saved", false))
}

func (d *Daemon) chanIdToName(id string) string {
	name, ok := d.chanIdToNameCache[id]
	if ok {
		return name
	}
	g, err := d.api.GetConversationInfo(id, false)
	if err != nil {
		log.Debugf("GetChannelInfo error: %s", err.Error())
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
		log.Errorf("GetUserInfo error: %s (You probably don't have the `users:read` scope)", err.Error())
		return "null"
	}
	d.userIdToNameCache[id] = u.Name
	return u.Name
}
