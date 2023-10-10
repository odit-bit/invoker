package index

import (
	"time"

	"github.com/google/uuid"
)

//indexer will implement by object that can index and search documents
// discovered by crawler

type Document struct {
	//its non-empty field .
	//this field connect a document with link that obtained from graph
	LinkID uuid.UUID
	URL    string

	// is a <title> element if point to HTML
	Title string

	//block of text that extracted when processing a Link
	Content string

	// document last indexed timestamp,
	// use UTC as location
	IndexedAt time.Time `db:"indexed_at"`

	//pageracnk score by pagerank calculator
	PageRank float64
}

type Indexer interface {
	// index will insert or update the index entry (doc)
	Index(doc *Document) error

	// Perform a lookup for a document by its ID
	Lookup(linkID uuid.UUID) (*Document, error)

	// Perform a full-text query and obtain an iterable list of results
	Search(query Query) (Iterator, error)

	// Update the PageRank score for a particular document
	UpdateScore(linkID uuid.UUID, score float64) error
}

// implement by object that can paginated the result
type Iterator interface {
	Close() error

	Next() bool

	Error() error

	Document() *Document

	TotalCount() uint64
}

// determine query type that support by indexer
type QueryType uint8

const (
	QueryTypeMatch = iota
	QueryTypePhrase
)

type Query struct {
	Type QueryType

	// search expression
	// stores the search query that's
	// entered by the end user.
	Expression string

	// number of search document skipped
	Offset uint64
}
