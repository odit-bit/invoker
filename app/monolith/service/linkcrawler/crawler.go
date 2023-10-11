package linkcrawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/crawler"
	"github.com/odit-bit/invoker/internal/metric"
	"github.com/odit-bit/invoker/internal/privnet"
	"github.com/odit-bit/invoker/linkgraph/graph"
	"github.com/odit-bit/invoker/partition"
	"github.com/odit-bit/invoker/textIndex/index"
)

//===========================================================service

type GraphAPI interface {
	UpsertLink(link *graph.Link) error
	// insert the new edge, the updated scenario will occure
	// if crawler will discovered another link from edge destination it will need updated
	UpsertEdge(edge *graph.Edge) error

	// RemoveStaleEdges removes any edge that originates from the specified
	// link ID and was updated before the specified timestamp.
	RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error

	//return link iterator to iterate link in graph
	Links(fromID, toID uuid.UUID, retrieveBefore time.Time) (graph.LinkIterator, error)
}

type IndexAPI interface {
	// index will insert or update the index entry (doc)
	Index(doc *index.Document) error
}

// metric
type CountLinkMetric interface {
	Add(float64)
}

// encapsulate component that service need
type Config struct {
	// managing links and edges in linkgraph
	Graphdb GraphAPI

	//indexing document
	Indexdb IndexAPI

	//perfroming HTTP request
	URLGetter crawler.URLGetter

	// detect private network address defined in RFC1918
	NetDetector crawler.PrivateNetworkDetector

	// wake the crawler to start scan the link again
	UpdateInterval time.Duration

	// minimum amount of time before re-indexing link that already crawled
	ReindexInterval time.Duration

	//detect partition assginment for this service
	PartitionDetector partition.Detector

	//number conccurent worker used for retreiving link.
	FetchWorker int

	// count amount of crawled link for this service
	Counter metric.CounterFunc

	// log event
	Logger *log.Logger
}

func (cfg *Config) validate() error {
	if cfg.Graphdb == nil {
		return fmt.Errorf("graph-db cannot nil")
	}

	if cfg.Indexdb == nil {
		return fmt.Errorf("index-db cannot nil")
	}

	if cfg.Counter == nil {
		cfg.Counter = func(_ float64) {}
	}

	if cfg.URLGetter == nil {
		return fmt.Errorf("urlGetter detector not been provided")
	}

	if cfg.FetchWorker == 0 {
		return fmt.Errorf("FetchWorker should more than 0")
	}

	if cfg.NetDetector == nil {
		return fmt.Errorf("FetchWorker should more than 0")
	}

	if cfg.PartitionDetector == nil {
		return fmt.Errorf("partition detector not been provided")
	}

	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stdout, "[crawler]", log.Ldate|log.Ltime)

	}

	return nil
}

type Service struct {
	cfg *Config

	//crawler pipeline
	crawler *crawler.Crawler
}

// crawl the url from default config
func New(graphDB GraphAPI, indexDB IndexAPI) *Service {
	partition := partition.Fixed{
		Partition:     0,
		NumPartitions: 1,
	}

	urlGetter := http.DefaultClient //xhttpclient.NewUrlGetterWithTimeout(2 * time.Second)

	Logger := log.New(os.Stdout, "[crawler]", log.Ldate|log.Ltime)

	netDetector, _ := privnet.NewDetector()

	conf := Config{
		Graphdb:           graphDB,
		Indexdb:           indexDB,
		URLGetter:         urlGetter,
		NetDetector:       netDetector,
		UpdateInterval:    1 * time.Minute,
		ReindexInterval:   7 * 24 * time.Hour,
		PartitionDetector: partition,
		FetchWorker:       1,
		Counter:           func(f float64) { _ = f },
		Logger:            Logger,
	}

	crawlService, err := NewWithConfig(&conf)
	if err != nil {
		log.Fatal(err)
	}

	return crawlService
}

func NewWithConfig(cfg *Config) (*Service, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// pipeline
	pipe, err := crawler.New(&crawler.Config{
		URLGetter:    cfg.URLGetter,
		NetDetector:  cfg.NetDetector,
		Indexer:      cfg.Indexdb,
		GraphUpdater: cfg.Graphdb,
		FetchWorker:  cfg.FetchWorker,
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:     cfg,
		crawler: pipe,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	s.cfg.Logger.Println("crawler service start")
	s.cfg.Logger.Printf("update interval: %v\n", s.cfg.UpdateInterval.String())
	s.cfg.Logger.Printf("reindex interval: %v\n", s.cfg.ReindexInterval.String())
	s.cfg.Logger.Printf("worker: %v\n", s.cfg.FetchWorker)

	ticker := time.NewTimer(s.cfg.UpdateInterval)
	defer func() {
		if ticker.Stop() {
			<-ticker.C
		}

	}()
	for {

		select {
		case <-ctx.Done():
			if !ticker.Stop() {
				<-ticker.C
			}
			return nil
		case <-ticker.C:
			s.cfg.Logger.Println("[INFO] crawl iteration start")
			cur, num, err := s.cfg.PartitionDetector.PartitionInfo()
			if err != nil {
				return err
			}
			if err := s.crawlGraph(ctx, cur, num); err != nil {
				return err
			}
			ticker.Reset(s.cfg.UpdateInterval)
		}
	}
}

// crawling
func (s *Service) crawlGraph(ctx context.Context, curPartition, numPartition int) error {
	r, err := partition.NewFullRange(numPartition)
	if err != nil {
		return err
	}

	fromID, toID, err := r.PartitionExtents(curPartition)
	if err != nil {
		return err
	}

	start := time.Now()
	li, err := s.cfg.Graphdb.Links(fromID, toID, start.Add(-s.cfg.ReindexInterval))
	if err != nil {
		return err
	}
	defer li.Close()

	n, err := s.crawler.Crawl(ctx, li)
	end := time.Since(start)
	if err != nil {
		return err
	}

	s.cfg.Counter(float64(n))
	s.cfg.Logger.Printf("[INFO] completed pipeline link:%v et: %v \n", n, end.Round(1*time.Millisecond))
	return nil
}
