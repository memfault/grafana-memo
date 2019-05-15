package daemon

import (
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/raintank/dur"
)

func ParseCommand(msg string, cl clock.Clock) (time.Time, string, []string, error) {
	msg = strings.TrimSpace(msg)
	words := strings.Fields(msg)
	if len(words) == 0 {
		return time.Time{}, "", nil, errEmpty
	}

	// parse time offset out of message (if applicable) and set timestamp
	ts := cl.Now().Add(-25 * time.Second)
	dur, err := dur.ParseDuration(words[0])
	if err == nil {
		ts = cl.Now().Add(-time.Duration(dur) * time.Second)
		words = words[1:]
	} else {
		parsed, err := time.Parse(time.RFC3339, words[0])
		if err == nil {
			ts = parsed
			words = words[1:]
		}
	}
	if len(words) == 0 {
		return time.Time{}, "", nil, errEmpty
	}

	// parse extra tags out of message (if applicable)
	pos := len(words) - 1 // pos of the last word that is not a tag
	for strings.Contains(words[pos], ":") {
		pos--
		if pos < 0 {
			return ts, msg, nil, errEmpty
		}
	}
	extraTags := words[pos+1:]

	return ts.UTC(), strings.Join(words[:pos+1], " "), extraTags, nil

}
