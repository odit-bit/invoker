package memory

import (
	"reflect"
	"testing"
	"time"

	"github.com/odit-bit/invoker/linkgraph/graph"
)

// func Test_Memory(t *testing.T) {
// 	t1 := time.Now().Add(-1 * time.Hour)
// 	l1 := &graph.Link{
// 		URL:         "www.example1.com",
// 		RetrievedAt: t1,
// 	}

// 	cache := New()

// 	err := cache.UpsertLink(l1)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t2 := time.Now().Add(-30 * time.Minute)
// 	l2 := &graph.Link{
// 		URL:         "www.example2.com",
// 		RetrievedAt: t2,
// 	}

// 	err = cache.UpsertLink(l2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	//Lookup the link
// 	rl1, err := cache.LookupLink(l1.ID)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if !reflect.DeepEqual(rl1, l1) {
// 		t.Errorf("not same\nactual:\t%v\nexpected:\t%v\n", rl1, l1)
// 	}

// 	//Lookup the link
// 	rl2, err := cache.LookupLink(l2.ID)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if !reflect.DeepEqual(rl1, l1) {
// 		t.Errorf("not same\nactual:\t%v\nexpected:\t%v\n", rl2, l2)
// 	}

// }

func Test_iterator(t *testing.T) {
	t1 := time.Now().Add(-1 * time.Hour)
	l1 := &graph.Link{
		URL:         "www.example1.com",
		RetrievedAt: t1,
	}

	t2 := time.Now().Add(-30 * time.Minute)
	l2 := &graph.Link{
		URL:         "www.example2.com",
		RetrievedAt: t2,
	}

	cache := New()
	cache.UpsertLink(l1)
	cache.UpsertLink(l2)

	// list
	list, err := cache.Links(l1.ID, l2.ID, time.Now())
	if err != nil {
		t.Error(err)
	}

	expectedCount := 1
	count := 0
	for list.Next() {
		count++

		val := list.Link()
		if !reflect.DeepEqual(val, l1) {
			t.Logf("not same\nactual:\t%v\nexpected:\t%v\n", val, l1)
			t.Fail()
		}
	}

	if count != expectedCount {
		t.Error()
	}

}
