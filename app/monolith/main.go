package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/odit-bit/invoker/app/monolith/service"
	"github.com/odit-bit/invoker/app/monolith/service/linkcrawler"
	"github.com/odit-bit/invoker/app/monolith/service/pagerank"
	"github.com/odit-bit/invoker/frontend"
	"github.com/odit-bit/invoker/internal/privnet"
	"github.com/odit-bit/invoker/internal/xhttpclient"
	"github.com/odit-bit/invoker/linkgraph/store/postgredb"
	"github.com/odit-bit/invoker/partition"
	"github.com/odit-bit/invoker/textIndex/store/postgreindex"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func connectPG(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed ping: %v", err)
	}

	return db, nil

}

func main() {
	var (
		pagerank_worker          int
		pagerank_update_interval time.Duration
	)

	var (
		crawler_worker           int
		crawler_update_interval  time.Duration
		crawler_reindex_interval time.Duration //time.Duration
	)

	var (
		dsn string
	)

	dur1, err := time.ParseDuration(os.Getenv("PAGERANK_UPDATE_TIME"))
	if err != nil {
		dur1 = 30 * time.Minute
	}

	dur2, err := time.ParseDuration(os.Getenv("CRAWLER_WAKE_INTERVAL"))
	if err != nil {
		dur2 = 5 * time.Minute
	}

	n2, err := strconv.Atoi(os.Getenv("CRAWLER_WORKER"))
	if err != nil {
		n2 = runtime.NumCPU()
	}

	// pagerank
	flag.IntVar(&pagerank_worker, "pagerank-worker", runtime.NumCPU(), "pagerank computing worker")
	flag.DurationVar(&pagerank_update_interval, "pagerank-update-interval time", dur1, "determined update rank time in minute")

	// crawler
	flag.IntVar(&crawler_worker, "crawler-worker ", n2, "crawler link fetcher worker")
	flag.DurationVar(&crawler_update_interval, "crawler-wake-interval", dur2, "determined wake crawler time in minute")
	flag.DurationVar(&crawler_reindex_interval, "crawler-reindex-treshold", 7*24*time.Hour, "determined time before link re-crawl again ")

	// dsn
	flag.StringVar(&dsn, "dsn ", os.Getenv("DSN"), "uri or string for data source (database)")
	flag.Parse()

	//=================

	if dsn == "" {
		dsn = "host=localhost user=development password=credential dbname=development sslmode=disable"
	}

	//======================persistence setup
	dbConn, err := connectPG(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	graphDB := postgredb.Newdb(dbConn)

	indexDB, err := postgreindex.New(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	//====================== Service
	// pagerank instance
	// pagerankService := pagerank.New(graphDB, indexDB)
	part := partition.Fixed{
		Partition:     0,
		NumPartitions: 1,
	}
	pagerankService, err := pagerank.NewWithConfig(pagerank.Config{
		GraphAPI:          graphDB,
		IndexAPI:          indexDB,
		PartitionDetector: part,
		ComputeWorkers:    pagerank_worker,
		UpdateInterval:    time.Duration(pagerank_update_interval),
	})
	if err != nil {
		log.Fatal(err)
	}

	// linkrawler instance
	urlGetter := xhttpclient.NewUrlGetterWithTimeout(3 * time.Second)
	detector, err := privnet.NewDetector()
	if err != nil {
		log.Fatal(err)
	}

	counter := promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawled_link_total",
		Help: "total crawled link",
	})

	// crawlService := linkcrawler.New(graphDB, indexDB)
	crawlService, err := linkcrawler.NewWithConfig(linkcrawler.Config{
		Graphdb:           graphDB,
		Indexdb:           indexDB,
		URLGetter:         urlGetter,
		NetDetector:       detector,
		UpdateInterval:    time.Duration(crawler_update_interval),
		ReindexInterval:   time.Duration(crawler_reindex_interval),
		PartitionDetector: part,
		FetchWorker:       crawler_worker,
		Counter:           counter.Add,
		Logger:            nil,
	})
	if err != nil {
		log.Fatal(err)
	}

	//frontend instance
	frontendService := frontend.NewDefault(graphDB, indexDB)

	var spv service.Supervised

	// run services
	spv = append(spv, pagerankService)
	spv = append(spv, crawlService)
	spv = append(spv, frontendService)

	//====================== start app

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	//gracefull shutdown
	go func() {
		// signal init
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP)

		//blocking wait signal
		select {

		case s := <-sigCh:
			log.Printf("signal :%v, info:shutdown \n", s.String())
			//this cancel func should make all services exit
			cancelFn()

		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				log.Printf("context : %v", err)
			}
		}

	}()

	if err := spv.Run(ctx); err != nil {
		log.Println(err)
	}

	log.Println("succesfull shutdown")
}
