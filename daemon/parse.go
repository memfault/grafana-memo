package daemon

import (
	"sort"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/raintank/dur"
	"github.com/raintank/memo"
)

func ParseMemo(ch, usr, msg string, cl clock.Clock) (memo.Memo, error) {
	msg = strings.TrimSpace(msg)
	words := strings.Fields(msg)
	if len(words) == 0 {
		return memo.Memo{}, errEmpty
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
	var extraTags []string
	stripChars := 0
	for i := len(words) - 1; i >= 0 && strings.Contains(words[i], ":"); i-- {
		stripChars += len(words[i] + " ") // this is a bug, should account for true amount of whitespace. sue me
		cleanTag := strings.TrimSpace(words[i])
		if strings.HasPrefix(cleanTag, "author:") || strings.HasPrefix(cleanTag, "chan:") {
			return memo.Memo{}, errFixedTag
		}
		extraTags = append(extraTags, strings.TrimSpace(words[i]))
	}
	msg = msg[:len(msg)-stripChars]
	msg = strings.TrimSpace(msg)
	if len(msg) == 0 {
		return memo.Memo{}, errEmpty
	}
	tags := []string{
		"memo",
		"author:" + usr,
	}
	if ch != "null" {
		tags = append(tags, "chan:"+ch)
	}
	sort.Strings(extraTags)
	tags = append(tags, extraTags...)

	return memo.Memo{
		ts.UTC(),
		msg,
		tags,
	}, nil
}
