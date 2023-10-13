package postgreindex

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/textIndex/index"
)

const lookupDocumentQuery = `
	SELECT linkID, url, title, content, indexed_at, pagerank FROM documents
	WHERE linkID = $1
`

// Lookup implements index.Indexer.
func (i *indexdb) Lookup(linkID uuid.UUID) (*index.Document, error) {
	var doc index.Document
	err := i.db.QueryRowxContext(context.TODO(), lookupDocumentQuery, linkID).Scan(
		&doc.LinkID,
		&doc.URL,
		&doc.Title,
		&doc.Content,
		&doc.IndexedAt,
		&doc.PageRank,
	)
	if err != nil {
		return nil, fmt.Errorf("indexer lookup document: %v", err)
	}
	return &doc, nil
}
