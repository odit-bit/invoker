package crawler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/odit-bit/invoker/internal/pipeline"
	"github.com/odit-bit/invoker/textIndex/index"
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
	if payload.Title == "" || payload.TextContent == "" {
		log.Println("title or content is nil, it will error on postgreindex, url:", payload.URL)
	}

	doc := index.Document{
		LinkID:    payload.LinkID,
		URL:       payload.URL,
		Title:     payload.Title,
		Content:   payload.TextContent,
		IndexedAt: time.Now(),
		PageRank:  0,
	}
	if err := ti.indexer.Index(&doc); err != nil {
		return nil, err
	}

	return p, nil

}
