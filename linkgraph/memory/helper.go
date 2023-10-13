package memory

import (
	"github.com/google/uuid"
	"github.com/odit-bit/invoker/linkgraph/graph"
)

// avoid collision to make sure the id is unique ,
// will iterate until uuid not exist in store.
// return unique uuid
func mustUniqueEdge(edgeStore map[uuid.UUID]*graph.Edge) uuid.UUID {
	var id uuid.UUID
	for {
		id = uuid.New()
		if edgeStore[id] == nil {
			break
		}
	}
	return id
}
