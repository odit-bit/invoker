package postgreindex

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/textIndex/index"
)

const insertDocumentQuery = `
	INSERT INTO documents (linkID, url, title, content, indexed_at, pagerank)
	VALUES($1,$2,$3,$4, $5, $6)
	ON CONFLICT (linkID) DO 
	UPDATE
		SET url = EXCLUDED.url,
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			indexed_at = NOW();
`

// Index implements index.Indexer.
// it uses to insert new document
func (i *indexdb) Index(doc *index.Document) error {

	if doc.LinkID == uuid.Nil {
		return fmt.Errorf("indexer insert document: uuid cannot be nil")
	}
	doc.IndexedAt = doc.IndexedAt.UTC()
	_, err := i.db.ExecContext(context.TODO(), insertDocumentQuery, doc.LinkID, doc.URL, doc.Title, doc.Content, doc.IndexedAt, doc.PageRank)
	if err != nil {
		log.Printf("indexer index error: title: %v \n", doc.Title)
		log.Printf("indexer index error: content: %v \n", doc.Content)
		log.Printf("indexer index error: content: %v \n", doc.IndexedAt)

		return fmt.Errorf("indexer insert document error: %v, doc detail: %v", err, doc.URL)
	}
	return nil
}

const updateScoreQuery = `
	UPDATE documents
	SET pagerank = $1 -- Replace with the pagerank value
	WHERE linkID = $2; -- Replace with the specific linkID 

`

// UpdateScore implements index.Indexer.
// update the pagerank score's dcoument by linkID
func (i *indexdb) UpdateScore(linkID uuid.UUID, score float64) error {
	_, err := i.db.ExecContext(context.TODO(), updateScoreQuery, score, linkID)
	if err != nil {
		return fmt.Errorf("update pagerank document : %v", err)
	}
	return nil
}
