package postgreindex

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/odit-bit/invoker/textIndex/index"
)

func Test_postgre_indexer(t *testing.T) {
	db, err := sqlx.Connect("pgx", "host=localhost user=development password=credential dbname=development sslmode=disable")
	if err != nil {
		t.Fatal("open db conn:", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal("try ping db:", err)
	}
	pgIndex, err := New(db)
	if err != nil {
		t.Fatal("create postgreindex instance:", err)
	}

	defer func() {
		err := pgIndex.drop()
		if err != nil {
			t.Fatal(err)
		}
		db.Close()
	}()

	// test itarator
	docs := createDoc(5)
	for _, doc := range docs {
		err := pgIndex.Index(&doc)
		if err != nil {
			t.Fatal(err)
		}
	}

	docIt, err := pgIndex.Search(index.Query{
		Type:       0,
		Expression: "example",
		Offset:     0,
	})

	if err != nil {
		t.Fatal(err)
	}
	defer docIt.Close()

	// assertDocIterator
	asserDocIterator(docs, docIt, t)

	//===================
	idx1 := &index.Document{
		LinkID:    uuid.New(),
		URL:       "www.example.com",
		Title:     "example",
		Content:   "content example",
		IndexedAt: time.Now().UTC(),
		PageRank:  0,
	}
	err = pgIndex.Index(idx1)

	if err != nil {
		t.Fatal(err)
	}

	idxRes, err := pgIndex.Lookup(idx1.LinkID)
	if err != nil {
		t.Fatal(err)
	}
	assertDoc(idx1, idxRes, t)

	err = pgIndex.UpdateScore(idx1.LinkID, 0.8)
	if err != nil {
		t.Fatal("update pagerank doc", err)
	}

	idx1, err = pgIndex.Lookup(idx1.LinkID)
	if err != nil {
		t.Fatal(err)
	}

	if idx1.PageRank != 0.8 {
		t.Fatal("failed update pager rank score", idx1.PageRank)
	}

}

func asserDocIterator(expect []index.Document, docIt index.Iterator, t *testing.T) {
	// sort the expected
	sort.Slice(expect, func(i, j int) bool {
		return expect[i].PageRank >= expect[j].PageRank
	})

	// check total count
	if total := docIt.TotalCount(); total != uint64(len(expect)) {
		t.Fatal("total count not meet expected got: ", total)
	}

	count := 0

	for docIt.Next() {
		doc := docIt.Document()
		if !bytes.Equal(expect[count].LinkID[:], doc.LinkID[:]) {
			t.Logf("different doc as expected\nexp:%v \ngot:%v \n", expect[count], doc)
			t.FailNow()
		}
		count++
	}
	if err := docIt.Error(); err != nil {
		t.Fatal(err)
	}

	if count == 0 {
		t.Fatal("iterator not iterate", count)
	}

}

func createDoc(n int) []index.Document {
	docs := []index.Document{}
	for i := 0; i < n; i++ {
		doc := index.Document{
			LinkID:    uuid.New(),
			URL:       fmt.Sprintf("www.example_%v.com", i),
			Title:     fmt.Sprintf("example_%v", i),
			Content:   fmt.Sprintf("content example_%v", i),
			IndexedAt: time.Now().UTC(),
			PageRank:  float64(i),
		}
		docs = append(docs, doc)
	}
	return docs
}

func assertDoc(idx1, idxRes *index.Document, t *testing.T) {
	if !bytes.Equal(idx1.LinkID[:], idxRes.LinkID[:]) {
		t.Fatal("lookup document")
	}

	if idx1.URL != idxRes.URL {
		t.Fatal("lookup document url")
	}

	if idx1.IndexedAt.Unix() != idxRes.IndexedAt.Unix() {
		t.Logf(
			"\nact:%v\nres:%v",
			idx1.IndexedAt.Unix(),
			idxRes.IndexedAt.Unix(),
		)
		t.Fatal("lookup document indexed_at")
	}
}
