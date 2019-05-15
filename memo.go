package memo

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrFixedTag = errors.New("cannot override built-in tags")

type Memo struct {
	Date time.Time
	Desc string
	Tags []string
}

// BuildTags takes the base tags (hardcoded), and extra tags (user specified)
// it validates the user is not trying to override the built in tags,
// merges them and sorts them
func BuildTags(base, extra []string) ([]string, error) {
	for _, ext := range extra {
		if strings.HasPrefix(ext, "author:") {
			return nil, ErrFixedTag
		}
		if strings.HasPrefix(ext, "chan:") {
			return nil, ErrFixedTag
		}
	}
	base = append(base, extra...)
	sort.Strings(base)
	return base, nil
}
