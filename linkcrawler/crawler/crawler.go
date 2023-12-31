package crawler

// this is the logic of crawling process,
// that use pipeline to divide crawling process into stage
// included defined processor for every stage

import (
	"context"
	"fmt"
	"sync"

	// "github.com/odit-bit/invoker/linkcrawler/pipeline"
	"github.com/odit-bit/invoker/linkgraph/graph"
	"github.com/odit-bit/pipeline"
)

var _ pipeline.Payload = (*payload)(nil)

var payloadPool = sync.Pool{
	New: func() any {
		return new(payload)
	},
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

var _ pipeline.Destination = (*countingSink)(nil)

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
	pipe *pipeline.Pipe
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

	stg1 := pipeline.NewMuxStage(cfg.FetchWorker, newLinkFetcher(cfg.URLGetter, cfg.NetDetector))
	stg2 := pipeline.NewFifo(newLinkExtractor(cfg.NetDetector))
	stg3 := pipeline.NewFifo(newTextExtractor())
	stg4 := pipeline.NewBroadcast(
		newUpdater(cfg.GraphUpdater),
		newTextIndexer(cfg.Indexer),
	)

	pipe := *pipeline.NewPipe(&pipeline.Config{},
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
	err := c.pipe.Run(ctx, &src, dst)
	return dst.getCount(), err
}
