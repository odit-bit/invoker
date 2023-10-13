package graphtest

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/linkgraph/graph"
	"github.com/odit-bit/invoker/linkgraph/memory"
)

// // SuiteBase defines a re-usable set of graph-related tests that can
// // be executed against any type that implements graph.Graph.
// type SuiteBase struct {
// 	g graph.Graph
// }

// // SetGraph configures the test-suite to run all tests against g.
// func (s *SuiteBase) SetGraph(g graph.Graph) {
// 	s.g = g
// }

// // TestUpsertLink verifies the link upsert logic.
// func (s *SuiteBase) TestUpsertLink(c *gc.C) {
// 	// Create a new link
// 	original := &graph.Link{
// 		URL:         "https://example.com",
// 		RetrievedAt: time.Now().Add(-10 * time.Hour),
// 	}

// 	err := s.g.UpsertLink(original)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(original.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a linkID to be assigned to the new link"))

// 	// Update existing link with a newer timestamp and different URL
// 	accessedAt := time.Now().Truncate(time.Second).UTC()
// 	existing := &graph.Link{
// 		ID:          original.ID,
// 		URL:         "https://example.com",
// 		RetrievedAt: accessedAt,
// 	}
// 	err = s.g.UpsertLink(existing)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(existing.ID, gc.Equals, original.ID, gc.Commentf("link ID changed while upserting"))

// 	stored, err := s.g.LookupLink(existing.ID)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(stored.RetrievedAt, gc.Equals, accessedAt, gc.Commentf("last accessed timestamp was not updated"))

// 	// Attempt to insert a new link whose URL matches an existing link with
// 	// and provide an older accessedAt value
// 	sameURL := &graph.Link{
// 		URL:         existing.URL,
// 		RetrievedAt: time.Now().Add(-10 * time.Hour).UTC(),
// 	}
// 	err = s.g.UpsertLink(sameURL)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(sameURL.ID, gc.Equals, existing.ID)

// 	stored, err = s.g.LookupLink(existing.ID)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(stored.RetrievedAt, gc.Equals, accessedAt, gc.Commentf("last accessed timestamp was overwritten with an older value"))

// 	// Create a new link and then attempt to update its URL to the same as
// 	// an existing link.
// 	dup := &graph.Link{
// 		URL: "foo",
// 	}
// 	err = s.g.UpsertLink(dup)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(dup.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a linkID to be assigned to the new link"))
// }

// Manual Testing

func assertErr(err error, msg string) func(t *testing.T) {
	return func(t *testing.T) {
		if err != nil {
			t.Error(err, msg)
		}
	}
}

// test UpsertLink method
func testUpsertLink(InMemory graph.Graph) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new link
		original := &graph.Link{
			URL:         "https://example.com",
			RetrievedAt: time.Now().Add(-10 * time.Hour),
		}

		err := InMemory.UpsertLink(original)
		if err != nil {
			t.Error(err)
		}
		//=============================
		// Update existing link with a newer timestamp and different URL
		accessedAt := time.Now().Truncate(time.Second).UTC()
		existing := &graph.Link{
			ID:          original.ID,
			URL:         "https://example.com",
			RetrievedAt: accessedAt,
		}

		err = InMemory.UpsertLink(existing)
		assertErr(err, "")(t)

		if existing.ID != original.ID {
			t.Errorf("\ngot:\t %v, \nexpected:\t %v \nerror: %v", existing.ID, original.ID,
				"value of ID should same")
		}
		stored, err := InMemory.LookupLink(existing.ID)
		assertErr(err, "")(t)
		if stored.RetrievedAt != accessedAt {
			t.Errorf("\ngot:\t %v, \nexpected:\t %v \nerror: %v", stored.RetrievedAt, accessedAt,
				"last accessed timestamp was not updated")
		}
		//=============================
		// Attempt to insert a new link whose URL matches an existing link with
		// and provide an older accessedAt value
		sameURL := &graph.Link{
			URL:         existing.URL,
			RetrievedAt: time.Now().Add(-10 * time.Hour).UTC(),
		}
		err = InMemory.UpsertLink(sameURL)
		assertErr(err, "")(t)

		if existing.ID != sameURL.ID {
			t.Errorf("\ngot:\t %v, \nexpected:\t %v \nerror: %v", existing.ID, sameURL.ID,
				"value of ID should same")
		}

		stored, err = InMemory.LookupLink(existing.ID)

		assertErr(err, "")(t)
		if stored.RetrievedAt != accessedAt {
			t.Errorf("\ngot:\t %v, \nexpected:\t %v \nerror: %v", stored.RetrievedAt, accessedAt,
				"last accessed timestamp was overwritten with an older value")
		}

		//=============================
		// Create a new link and then attempt to update its URL to the same as
		// an existing link.
		dup := &graph.Link{
			URL: "foo",
		}
		err = InMemory.UpsertLink(dup)
		assertErr(err, "")
		// c.Assert(dup.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a linkID to be assigned to the new link"))
		if dup.ID == uuid.Nil {
			t.Errorf("\ngot:\t %v, \nexpected:\t %v \nerror: %v", dup.ID, uuid.Nil,
				"last accessed timestamp was overwritten with an older value")
		}
	}
}

func Test_UpsertLink(t *testing.T) {
	inMem := memory.New()
	t.Run("UpsertLink", testUpsertLink(inMem))
}

// // TestLinkIteratorTimeFilter verifies that the time-based filtering of the
// // link iterator works as expected.
// func testLinksIterator(InMemory graph.Graph) func(t *testing.T) {
// 	linkUUIDs := make([]uuid.UUID, 3)
// 	linkInsertTimes := make([]time.Time, len(linkUUIDs))

// 	return func(t *testing.T) {
// 		for i := 0; i < len(linkUUIDs); i++ {
// 			link := &graph.Link{URL: fmt.Sprint(i), RetrievedAt: time.Now()}
// 			err := InMemory.UpsertLink(link)
// 			assertErr(err, "")(t)
// 			linkUUIDs[i] = link.ID
// 			linkInsertTimes[i] = time.Now()
// 		}

// 		for i, t := range linkInsertTimes {

// 		}
// 	}
// }
