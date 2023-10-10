package memory

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/odit-bit/invoker/textIndex/index"
)

//custom indexer iterator implementation

var _ index.Iterator = (*IndexIterator)(nil)

type IndexIterator struct {
	//access in-memory stored document
	store *bleveMemory

	// last fetched doc from store
	lastDoc *index.Document
	// last error if any
	lastErr error

	// provide information about both the total number of matched results and
	// the number of documents in the current result batch.
	// because it would be matched doc exist is more than the current size of result,
	// if matched doc is 100 and the result size 10, theres is 90 be more doc available
	// and ready to fetch
	result *bleve.SearchResult

	//track the current position of (offset) batch from result
	resultIdx int

	// track the current number of matched doc that
	// not yet fetch is available
	globalIdx uint64

	//need to fetch the next batch
	request *bleve.SearchRequest
}

// Close implements index.Iterator.
func (it *IndexIterator) Close() error {
	it.store = nil
	it.request = nil
	if it.result != nil {
		it.globalIdx = it.result.Total
	}
	return nil
}

// Document implements index.Iterator.
func (it *IndexIterator) Document() *index.Document {
	return it.lastDoc
}

// Error implements index.Iterator.
func (it *IndexIterator) Error() error {
	return it.lastErr
}

// Next implements index.Iterator.
func (it *IndexIterator) Next() bool {
	// in sake of readability

	// there is no err and result exist
	if it.lastErr != nil || it.result == nil {
		return false
	}

	// check result is not exhausted
	if it.globalIdx >= it.result.Total {
		return false
	}

	// indicate need to fecthed the next batch ??
	if it.resultIdx >= it.result.Hits.Len() {
		//request doc per page size
		it.request.From += it.request.Size
		if it.result, it.lastErr = it.store.idx.Search(it.request); it.lastErr != nil {
			return false
		}

		// reset current page idx
		it.resultIdx = 0
	}

	// get next id of doc from result
	nextID := it.result.Hits[it.resultIdx].ID
	it.lastDoc, it.lastErr = it.store.lookupUUIDString(nextID)
	if it.lastErr != nil {
		return false
	}

	it.resultIdx++
	it.globalIdx++
	return true
}

// TotalCount implements index.Iterator.
func (it *IndexIterator) TotalCount() uint64 {
	if it.result == nil {
		return 0
	}
	return it.result.Total
}
