package store

import (
	"errors"

	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/grafana/memo"
)

// Elastic is an annotations store backed by elasticsearch
// it's hardcoded to use localhost, and deprecated
type Elastic struct {
	conn *elastigo.Conn
}

func NewElastic() Elastic {
	var e Elastic
	e.conn = elastigo.NewConn()
	e.conn.Domain = "localhost"
	e.conn.Port = "9200"
	return e
}

func (e Elastic) Save(memo memo.Memo) error {
	response, err := e.conn.Index("memos", "memo", "", nil, memo)
	if err != nil {
		return err
	}
	if !response.Created {
		return errors.New("failure to save to elasticsearch for unknown reason")
	}
	return nil
}
