package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/internal/pipeline"
	"github.com/odit-bit/invoker/linkgraph/graph"
)

var _ pipeline.Payload = (*payload)(nil)

var payloadPool = sync.Pool{
	New: func() any {
		return new(payload)
	},
}

// payload implemet pipeline.Payload
// so that can use by pipeline.Processor
type payload struct {
	// populated by input source
	LinkID      uuid.UUID
	URL         string
	RetrievedAt time.Time

	NoFollowLinks []string
	RawContent    bytes.Buffer

	Links       []string
	Title       string
	TextContent string
}

// Clone implements pipeline.Payload.
func (p *payload) Clone() pipeline.Payload {
	cloneP := payloadPool.Get().(*payload)
	cloneP.LinkID = p.LinkID
	cloneP.URL = p.URL
	cloneP.RetrievedAt = p.RetrievedAt
	cloneP.NoFollowLinks = append([]string(nil), p.NoFollowLinks...)
	cloneP.Links = append([]string(nil), p.Links...)
	cloneP.Title = p.Title
	cloneP.TextContent = p.TextContent

	_, err := io.Copy(&cloneP.RawContent, &p.RawContent)
	if err != nil {
		panic(fmt.Sprintf("[BUG] error cloning payload raw content: %v", err))
	}
	return cloneP
}

// MarkAsProcessed implements pipeline.Payload.
func (p *payload) MarkAsProcessed() {
	p.URL = p.URL[:0]
	p.Links = p.Links[:0]
	p.NoFollowLinks = p.NoFollowLinks[:0]
	p.Title = p.Title[:0]
	p.TextContent = p.TextContent[:0]
	p.RawContent.Reset()
	payloadPool.Put(p)
}

var _ pipeline.Source = (*LinkSource)(nil)

// poppulate link from graph as source of pipeline
type LinkSource struct {
	linkIter graph.LinkIterator
}

// Error implements pipeline.Source.
func (ls *LinkSource) Error() error {
	return ls.linkIter.Error()
}

// Next implements pipeline.Source.
func (ls *LinkSource) Next() bool {
	return ls.linkIter.Next()
}

// Payload implements pipeline.Source.
func (ls *LinkSource) Payload() pipeline.Payload {
	link := ls.linkIter.Link()

	p := payloadPool.Get().(*payload)
	p.LinkID = link.ID
	p.URL = link.URL
	p.RetrievedAt = link.RetrievedAt

	return p
}

var _ pipeline.Sink = (*countingSink)(nil)

type countingSink struct {
	count int
}

func (s *countingSink) Consume(_ context.Context, p pipeline.Payload) error {
	s.count++
	return nil
}

func (s *countingSink) getCount() int {
	// The broadcast split-stage sends out two payloads for each incoming link
	// so we need to divide the total count by 2.
	return s.count / 2
}

// crawler

// encapsulate options to create new Crawler
type Config struct {
	URLGetter    URLGetter
	NetDetector  PrivateNetworkDetector
	Indexer      Indexer
	GraphUpdater GraphUpdater

	FetchWorker int
}

func (c *Config) validate() error {
	if c.GraphUpdater == nil {
		return fmt.Errorf("graphUpdater not been provided")
	}
	if c.Indexer == nil {
		return fmt.Errorf("indexer not been provided")
	}
	if c.FetchWorker == 0 {
		return fmt.Errorf("invalide fetch worker value %v", c.FetchWorker)
	}
	if c.URLGetter == nil {
		return fmt.Errorf("urlGetter not been provided")
	}

	if c.NetDetector == nil {
		return fmt.Errorf("netDetector not been provided")
	}
	return nil
}

type Crawler struct {
	pipe *pipeline.Pipeline
}

// Crawler implements a web-page crawling pipeline consisting of the following
// stages:
//
//   - Given a URL, retrieve the web-page contents from the remote server.
//   - Extract and resolve absolute and relative links from the retrieved page.
//   - Extract page title and text content from the retrieved page.
//   - Update the link graph: add new links and create edges between the crawled
//     page and the links within it.
//   - Index crawled page title and text content.
func New(cfg *Config) (*Crawler, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	//stage 1
	stg1 := pipeline.WorkerPool(
		newLinkFetcher(cfg.URLGetter, cfg.NetDetector),
		cfg.FetchWorker,
	)

	stg2 := pipeline.FIFO(newLinkExtractor(cfg.NetDetector))
	stg3 := pipeline.FIFO(newTextExtractor())
	stg4 := pipeline.Broadcast(
		newUpdater(cfg.GraphUpdater),
		newTextIndexer(cfg.Indexer),
	)

	pipe := *pipeline.New(
		stg1,
		stg2,
		stg3,
		stg4,
	)
	return &Crawler{
		pipe: &pipe,
	}, nil

}

func (c *Crawler) Crawl(ctx context.Context, linkIterator graph.LinkIterator) (int, error) {
	src := LinkSource{
		linkIter: linkIterator,
	}

	dst := new(countingSink)
	err := c.pipe.Process(ctx, &src, dst)
	return dst.getCount(), err
}
