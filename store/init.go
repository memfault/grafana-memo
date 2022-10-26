package store

import "github.com/grafana/memo"

// Store
type Store interface {
	// Save stores the memo in the storage engine
	Save(memo memo.Memo) error
}
