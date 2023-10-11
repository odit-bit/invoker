package crawler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/internal/pipeline"
	"github.com/odit-bit/invoker/linkgraph/graph"
)

var _ pipeline.Processor = (*updater)(nil)

// will update payload into graph
type updater struct {
	graphUpdater GraphUpdater
}

func newUpdater(gu GraphUpdater) *updater {
	return &updater{
		graphUpdater: gu,
	}
}

// Process implements pipeline.Processor.
func (u *updater) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload, ok := p.(*payload)
	if !ok {
		return nil, fmt.Errorf("payload underlying type is not crawler's payload: %t", p)
	}
	// upsert link
	linkSrc := &graph.Link{
		ID:          payload.LinkID,
		URL:         payload.URL,
		RetrievedAt: time.Now(),
	}
	err := u.graphUpdater.UpsertLink(linkSrc)
	if err != nil {
		return nil, err
	}

	// insert nofollow link, without create an edge
	for _, dstLink := range payload.NoFollowLinks {
		l := graph.Link{URL: dstLink}
		err := u.graphUpdater.UpsertLink(&l)
		if err != nil {
			return nil, err
		}
	}

	//

	for _, dstLink := range payload.Links {
		dst := &graph.Link{
			URL: dstLink,
		}
		//insert link to follow
		err := u.graphUpdater.UpsertLink(dst)
		if err != nil {
			return nil, err
		}

		// insert edge dst with l, and src with linkSrc
		e := &graph.Edge{
			Src: linkSrc.ID,
			Dst: dst.ID,
		}
		err = u.graphUpdater.UpsertEdge(e)
		if err != nil {
			return nil, err
		}

	}

	removeEdgeBefore := time.Now()
	err = u.graphUpdater.RemoveStaleEdges(linkSrc.ID, removeEdgeBefore)
	if err != nil {
		return nil, err
	}

	return p, nil

}

// a list methods needed for the updater to communicate with a link
// graph component
type GraphUpdater interface {
	UpsertLink(link *graph.Link) error
	UpsertEdge(edge *graph.Edge) error
	RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error
}
