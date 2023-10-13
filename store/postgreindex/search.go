package postgreindex

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/odit-bit/invoker/textIndex/index"
)

// like matchDocQuery but whe $1 is blank or space it will return all
const searchDocCountQuery = `
SELECT COUNT(*) FROM documents
WHERE
	CASE
		WHEN length(trim($1)) = 0 THEN true
		ELSE ts @@ websearch_to_tsquery('english', $1)
	END
`

// plainto_tsquery vs to_tsquery
var searchDocQuery = `
SELECT linkID, url, title, content, indexed_at, pagerank
FROM documents
WHERE
	CASE
		WHEN length(trim($1)) = 0 THEN true
		ELSE ts @@ websearch_to_tsquery('english', $1)
	END
ORDER BY
	pagerank DESC,

	CASE
		WHEN length(trim($1)) = 0 THEN NULL
	 	ELSE ts_rank(ts, websearch_to_tsquery('english', $1), 32)
	END DESC	

OFFSET ($2) ROWS
FETCH FIRST ($3) ROWS ONLY;
`

// Search full-text index document.
func (i *indexdb) Search(query index.Query) (index.Iterator, error) {
	//get the matchedCount document
	var matchedCount int
	err := i.db.QueryRowxContext(context.TODO(), searchDocCountQuery, query.Expression).Scan(&matchedCount)
	if err != nil {
		return nil, fmt.Errorf("index search documents matched count: %v", err)
	}

	pageSize := batchSize
	offset := query.Offset

	rows, err := i.db.QueryxContext(context.TODO(), searchDocQuery, query.Expression, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("index search documents: %v", err)
	}

	docIterator := iterator{
		rows:         rows,
		latchedDoc:   nil,
		latchedErr:   nil,
		expression:   query.Expression,
		totalMatched: matchedCount,
	}
	return &docIterator, err
}

var _ index.Iterator = (*iterator)(nil)

type iterator struct {
	rows *sqlx.Rows
	// fetch      *sqlx.Stmt
	latchedDoc *index.Document
	latchedErr error

	expression   string
	totalMatched int
}

// Close implements index.Iterator.
func (it *iterator) Close() error {
	err := it.rows.Close()
	if err != nil {
		return err
	}

	return nil

}

// Document implements index.Iterator.
func (it *iterator) Document() *index.Document {
	return it.latchedDoc
}

// Error implements index.Iterator.
func (it *iterator) Error() error {
	return it.latchedErr
}

// Next implements index.Iterator.
func (it *iterator) Next() bool {
	return it.fecthDoc()

}

// TotalCount implements index.Iterator.
func (it *iterator) TotalCount() uint64 {
	return uint64(it.totalMatched)
}

func (it *iterator) fecthDoc() bool {
	ok := it.rows.Next()

	if !ok {
		it.latchedErr = it.rows.Err()
		return false
	}

	var doc index.Document
	err := it.rows.Scan(
		&doc.LinkID,
		&doc.URL,
		&doc.Title,
		&doc.Content,
		&doc.IndexedAt,
		&doc.PageRank,
	)
	if err != nil {
		it.latchedErr = err
		return false
	}
	it.latchedDoc = &doc
	return true
}
