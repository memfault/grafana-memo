package store

import "github.com/raintank/memo"

type Store interface {
	Save(memo memo.Memo) error
}
