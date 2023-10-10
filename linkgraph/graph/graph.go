package graph

import (
	"time"

	"github.com/google/uuid"
)

// Link represent the set of web pages that have been processed or discovered
// by the crawler component
type Link struct {
	// unique identifier for link
	ID uuid.UUID

	// link target
	URL string `db:"url"`

	// timestamp when link retrieved after processed
	RetrievedAt time.Time `db:"retrieved_at"`
}

// Edge represents a uni-directional connection between two links in the graph.
// describe graph edge that originate from src and endup at dst,
type Edge struct {
	//unique identifier
	ID uuid.UUID

	Src uuid.UUID //Link ID

	Dst uuid.UUID // Link ID

	// timestamp when link is update
	UpdateAt time.Time `db:"update_at"`
}

//defined the graph operation
/*
1. insert Link into graph or update existing link
2. lookup link by its ID
3. iterate all link presented in graph
4. insert or updated edge into graph
5. iterate the list edges in graph. it REQUIRED By PageRank calculator
6. Delete link that not updated bye crawler
*/
// because Graph is an abstract data type , so it make sense to defined upfront,

type Graph interface {
	//
	UpsertLink(link *Link) error

	//
	LookupLink(id uuid.UUID) (*Link, error)

	//return link iterator to iterate link in graph
	Links(fromID, toID uuid.UUID, retrieveBefore time.Time) (LinkIterator, error)

	// insert the new edge, the updated scenario will occure
	// if crawler will discovered another link from edge destination it will need updated
	UpsertEdge(edge *Edge) error

	// LookupEdge(id uuid.UUID) (*Edge, error)

	Edges(fromID, toID uuid.UUID, updateBefore time.Time) (EdgeIterator, error)

	// RemoveStaleEdges removes any edge that originates from the specified
	// link ID and was updated before the specified timestamp.
	RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error
}

// implemented by graph object that can be iterated
// the implementation detail is depend on underlying database technology
type Iterator interface {
	//advance the iterator , if not more item return false
	Next() bool

	// return last error encounterd by iterator
	Error() error

	//close release any resource associated with iterator
	Close() error
}

type LinkIterator interface {
	Iterator

	//return currently fetched Link
	Link() *Link
}

type EdgeIterator interface {
	Iterator

	// return currently fetched Edge
	Edge() *Edge
}
