package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/google/uuid"
	"github.com/odit-bit/invoker/textIndex/index"
)

//This package is about implementing document indexing and storing it in memory

// page size cache localy by iterator
const bacthSize = 10

// represent Implementation in-memory store indexer document
type bleveDoc struct {
	Title    string
	Content  string
	PageRank float64
}

var _ index.Indexer = (*bleveMemory)(nil)

// represent of in-memory store
type bleveMemory struct {
	mu sync.RWMutex

	// store
	docs map[string]*index.Document

	idx bleve.Index
}

func NewInMemoryIndexer() (*bleveMemory, error) {
	mapping := bleve.NewIndexMapping()
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	bv := &bleveMemory{
		mu:   sync.RWMutex{},
		docs: map[string]*index.Document{},
		idx:  idx,
	}

	return bv, nil
}

func (bm *bleveMemory) Close() error {
	return bm.idx.Close()
}

// Index implements index.Indexer.
func (bm *bleveMemory) Index(inputDoc *index.Document) error {

	if inputDoc.LinkID == uuid.Nil {
		return fmt.Errorf("missing doc Link ID")
	}

	inputDoc.IndexedAt = time.Now()
	dCopy := new(index.Document)
	*dCopy = *inputDoc

	key := dCopy.LinkID.String()

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// preserve existing PageRank score
	if origin, ok := bm.docs[key]; ok {
		dCopy.PageRank = origin.PageRank
	}

	//store to bleve index as bleve document
	//so no need to store the (maybe) big content
	err := bm.idx.Index(key, bleveDoc{
		Title:    dCopy.Title,
		Content:  dCopy.Content,
		PageRank: dCopy.PageRank,
	})
	if err != nil {
		return err
	}

	bm.docs[key] = dCopy
	return nil
}

// Lookup implements index.Indexer.
func (bm *bleveMemory) Lookup(linkID uuid.UUID) (*index.Document, error) {
	return bm.lookupUUIDString(linkID.String())
}

func (bm *bleveMemory) lookupUUIDString(linkID string) (*index.Document, error) {
	if doc, found := bm.docs[linkID]; found {
		return doc, nil
	}

	return nil, fmt.Errorf("lookup by ID : not found")
}

// Search implements index.Indexer.
func (bm *bleveMemory) Search(q index.Query) (index.Iterator, error) {
	var query query.Query
	switch q.Type {
	case index.QueryTypePhrase:
		query = bleve.NewMatchPhraseQuery(q.Expression)
	default:
		query = bleve.NewMatchQuery(q.Expression)
	}

	sr := bleve.NewSearchRequest(query)
	sr.SortBy([]string{"PageRank", "-_score"})
	sr.Size = bacthSize
	sr.From = int(q.Offset)

	rs, err := bm.idx.Search(sr)

	if err != nil {
		return nil, fmt.Errorf("index search: match not found")
	}
	iterator := &IndexIterator{
		store:     bm,
		lastErr:   err,
		result:    rs,
		globalIdx: q.Offset,
		request:   sr,
	}
	return iterator, nil
}

// UpdateScore implements index.Indexer.
func (bm *bleveMemory) UpdateScore(linkID uuid.UUID, score float64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	key := linkID.String()
	doc, found := bm.docs[key]
	if !found {
		doc = new(index.Document)
		doc.LinkID = linkID
		bm.docs[key] = doc
	}

	doc.PageRank = score
	err := bm.idx.Index(key, bleveDoc{
		Title:    doc.Title,
		Content:  doc.Content,
		PageRank: doc.PageRank,
	})
	if err != nil {
		return fmt.Errorf("update score: %v", err)
	}

	return nil

}
