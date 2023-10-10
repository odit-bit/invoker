package memory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/textIndex/index"
)

func Test_(t *testing.T) {
	doc := index.Document{
		LinkID:  uuid.New(),
		URL:     "www.example.com",
		Title:   "example",
		Content: "content",
	}

	c, err := NewInMemoryIndexer()
	if err != nil {
		t.Error(err)
	}

	err = c.Index(&doc)
	if err != nil {
		t.Error(err)
	}

	res, err := c.Search(index.Query{
		Type:       index.QueryTypeMatch,
		Expression: "example",
		Offset:     0,
	})

	if err != nil {
		t.Error(err)
	}

	if res.TotalCount() != 1 {
		t.Error(res.TotalCount())
	}

	count := 0
	for res.Next() {
		count++
		actual := res.Document()
		if actual.LinkID != doc.LinkID {
			t.Errorf("\ngot %v\n\nexpect %v\n", actual.LinkID, doc.LinkID)
		}
	}

	if count != 1 {
		t.Error(count)
	}
}
