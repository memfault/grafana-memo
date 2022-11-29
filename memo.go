package memo

import (
	"errors"
	"sort"
	"time"
)

// ErrEmpty used to return consistent error message for empty memo
var ErrEmpty = errors.New("empty message")

// HelpMessage used to return consistent help message
var HelpMessage = "Hi. I only support memo requests. See https://github.com/grafana/memo/blob/master/README.md#message-format"

// Memo
type Memo struct {
	// Date
	Date time.Time
	// Desc
	Desc string
	// Tags
	Tags []string
}

// BuildTags takes the base tags (hardcoded), and extra tags (user specified)
// it validates the user is not trying to override the built in tags,
// merges them and sorts them
func (m *Memo) BuildTags(extra []string) {
	base := []string{
		"memo",
	}

	base = append(base, extra...)
	sort.Strings(base)

	m.Tags = base
}
