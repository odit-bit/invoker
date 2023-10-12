package postgreindex

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/odit-bit/invoker/textIndex/index"
)

// it is like page-size
var batchSize int = 10

var _ index.Indexer = (*indexdb)(nil)

type indexdb struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) (*indexdb, error) {
	idb := indexdb{
		db: db,
	}

	err := idb.migrate()
	if err != nil {
		return nil, fmt.Errorf("postgreindex migrate: %v", err)
	}

	return &idb, nil

}
