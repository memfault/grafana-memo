package memo

import "time"

type Memo struct {
	Date time.Time
	Desc string
	Tags []string
}
