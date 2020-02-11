package store

import "github.com/grafana/memo"

type Store interface {
	Save(memo memo.Memo) error
}
