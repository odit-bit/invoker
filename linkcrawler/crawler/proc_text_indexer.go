package crawler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/odit-bit/invoker/textIndex/index"
	"github.com/odit-bit/pipeline"
)

var _ pipeline.Processor = (*textIndexer)(nil)

type Indexer interface {
	Index(doc *index.Document) error
}

type textIndexer struct {
	indexer Indexer
}

func newTextIndexer(ti Indexer) *textIndexer {
	return &textIndexer{
		indexer: ti,
	}
}

// Process implements pipeline.Processor.
func (ti *textIndexer) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload, ok := p.(*payload)
	if !ok {
		return nil, fmt.Errorf("graph updater not craweler's payload: %t ", p)
	}

	if len(payload.TextContent) == 0 {
		log.Println("[DEBUG][text indexer] text content is nil, it will error on postgreindex, url:", payload.URL)
	}

	doc := index.Document{
		LinkID:    payload.LinkID,
		URL:       payload.URL,
		Title:     string(payload.Title),
		Content:   string(payload.TextContent),
		IndexedAt: time.Now(),
		PageRank:  0,
	}
	if err := ti.indexer.Index(&doc); err != nil {
		return nil, err
	}

	return p, nil

}
