package parser

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/grafana/memo"
	"github.com/raintank/dur"

	log "github.com/sirupsen/logrus"
)

// Parser
type Parser struct {
	// re
	re *regexp.Regexp
}

// Parse takes a message and returns a memo with the fields extracted
func (p *Parser) Parse(message string) (*memo.Memo, error) {
	message = strings.TrimSpace(message)

	if len(message) == 0 {
		return nil, errors.New("message is empty")
	}

	memo := memo.Memo{}

	ok, err := p.isForUs(message)

	// regex match fail
	if err != nil {
		return nil, err
	}

	// not for us
	if !ok {
		return nil, nil
	}

	words := strings.Fields(message)
	if len(words) == 0 {
		return nil, errors.New("message is empty")
	}

	words, ts := p.extractTimestamp(words)
	memo.Date = ts
	memo.Desc = strings.Join(words, " ")

	pos := len(words) - 1 // pos of the last word that is not a tag
	for strings.Contains(words[pos], ":") {
		pos--
		if pos < 0 {
			return &memo, nil
		}
	}

	extraTags := words[pos+1:]
	memo.BuildTags(extraTags)

	memo.Desc = strings.Join(words[:pos+1], " ")

	return &memo, nil
}

// isForUs returns if this message has been identified as a memo
func (p *Parser) isForUs(message string) (bool, error) {
	out := p.re.FindStringSubmatch(message)
	if len(out) == 0 {
		if strings.HasPrefix(message, "memo:") || strings.HasPrefix(message, "mrbot:") || strings.HasPrefix(message, "memobot:") {
			log.Debugf("A user seems to direct a message `%q` to us, but we don't understand it. so sending help message back", message)
			return false, errors.New("message could not be understood")
		}

		// we're in a channel. don't spam in it. the message was probably not meant for us.
		log.Tracef("Received message `%q`, not for us. ignoring", message)
		return false, nil
	}

	return true, nil
}

// extractTimestamp takes a timestamp at the start of the memo
// written in RFC3339 format or time strings compatible with
// [https://pkg.go.dev/github.com/raintank/dur#ParseDuration]
func (p *Parser) extractTimestamp(words []string) ([]string, time.Time) {
	// parse time offset out of message (if applicable) and set timestamp
	ts := time.Now().Add(-25 * time.Second)
	dur, err := dur.ParseDuration(words[0])
	if err == nil {
		ts = time.Now().Add(-time.Duration(dur) * time.Second)
		words = words[1:]
	} else {
		parsed, err := time.Parse(time.RFC3339, words[0])
		if err == nil {
			ts = parsed
			words = words[1:]
		}
	}

	return words, ts
}

// New returns a new instance of Parser
func New() Parser {
	return Parser{
		re: regexp.MustCompile("^memo (.*)"),
	}
}
