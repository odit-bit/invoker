package crawler

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/textIndex/index"
)

var _ Indexer = (*mockIndexer)(nil)

type mockIndexer struct{}

// Index implements Indexer.
func (mi *mockIndexer) Index(doc *index.Document) error {
	return nil
}

func Test_indexer(t *testing.T) {

	idx := newTextIndexer(&mockIndexer{})
	_, err := idx.Process(context.Background(), &payload{
		LinkID:        uuid.New(),
		URL:           "",
		RetrievedAt:   time.Time{},
		NoFollowLinks: []string{},
		RawContent:    bytes.Buffer{},
		Links:         []string{},
		Title:         "",
		TextContent:   "",
	})

	if err != nil {
		t.Error(err)
	}
}
