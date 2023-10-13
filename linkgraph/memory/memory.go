package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/linkgraph/graph"
)

//represent Implementation in-memory store graph

// var _ graph.Graph = (*InMemory)(nil)

type InMemory struct {
	mu sync.RWMutex

	//core store implementation
	links map[uuid.UUID]*graph.Link
	edges map[uuid.UUID]*graph.Edge

	//use link url as key.
	//as index to speed retrieval if url is known
	linkUrlIndex map[string]*graph.Link

	/*
		use link ID as key, and list of edgeID as value
		associates link IDs with a slice of edge IDs.
		so value containt the list of
		edge that originate to same link (edge.Src)
	*/
	linkEdgeMap map[uuid.UUID]edgeList
}

// containt only the list of edge's ID that originate from the same link
/*
	example:
	[]edgelist{edge1,edge2}

	edge1.Src = www.example.com
	edge1.Dst = www.destination_1.com

	edge2.Src = www.example.com
	edge2.Dst = www.destination_2.com

*/
type edgeList []uuid.UUID // edge ID

func New() *InMemory {
	in := &InMemory{
		// mu:           sync.RWMutex{},
		links:        map[uuid.UUID]*graph.Link{},
		edges:        map[uuid.UUID]*graph.Edge{},
		linkUrlIndex: map[string]*graph.Link{},
		linkEdgeMap:  map[uuid.UUID]edgeList{},
	}

	return in
}

// RemoveStaleEdges implements graph.Graph.
func (in *InMemory) RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error {
	in.mu.Lock()
	defer in.mu.Unlock()

	var newEdgeList edgeList
	for _, edgeID := range in.linkEdgeMap[fromID] {
		edge := in.edges[edgeID]
		if edge.UpdateAt.Before(updatedBefore) {
			delete(in.edges, edgeID)
			continue
		}

		newEdgeList = append(newEdgeList, edgeID)
	}

	// Replace edge list or origin link with the filtered edge list
	in.linkEdgeMap[fromID] = newEdgeList
	return nil
}

func (in *InMemory) UpsertLink(link *graph.Link) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	// check if link is exist
	// is link exist update the link.ID into exist link
	if exist := in.linkUrlIndex[link.URL]; exist != nil {
		link.ID = exist.ID
		originRetrieve := exist.RetrievedAt
		*exist = *link

		if originRetrieve.After(exist.RetrievedAt) {
			exist.RetrievedAt = originRetrieve
		}
		return nil
	}

	// insert.
	// avoid collision uuid
	for {
		link.ID = uuid.New()
		if in.links[link.ID] == nil {
			break
		}
	}

	lcopy := new(graph.Link)
	*lcopy = *link
	//insert url as index
	in.linkUrlIndex[lcopy.URL] = lcopy
	//insert to stores
	in.links[lcopy.ID] = lcopy

	// log.Println("DEBUG inMEMORY LINKGRAPH links length:", len(in.links), lcopy.URL)
	return nil
}

// lookup link by it's ID
func (in *InMemory) LookupLink(id uuid.UUID) (*graph.Link, error) {
	in.mu.RLock()
	defer in.mu.RUnlock()

	l, ok := in.links[id]
	if !ok {
		return nil, fmt.Errorf("find link: not found ")
	}

	lcopy := new(graph.Link)
	*lcopy = *l
	return lcopy, nil
}

// return the set of links (iterator) whose id belong to
func (in *InMemory) Links(fromID, toID uuid.UUID, retrieveBefore time.Time) (graph.LinkIterator, error) {
	from, to := fromID.String(), toID.String()

	in.mu.RLock()
	var list []*graph.Link
	for linkID, link := range in.links {
		if id := linkID.String(); id >= from && id < to && link.RetrievedAt.Before(retrieveBefore) {
			list = append(list, link)
		}
	}
	in.mu.RUnlock()
	return &LinkIterator{
		s:    in,
		list: list,
		idx:  0,
	}, nil
}

// UpsertEdge implements graph.Graph.
func (in *InMemory) UpsertEdge(input *graph.Edge) error {
	in.mu.Lock()
	defer in.mu.Unlock()

	// 	verify that the source and destination links for the
	// edge actually exist.
	_, srcExist := in.links[input.Src]
	_, dstExist := in.links[input.Dst]

	if !srcExist || !dstExist {
		return fmt.Errorf("upsert edge: unknown edge link")
	}

	// update the edge
	//scan edge list from source to ensure the updatedAt time sync
	//with value contained in the store

	// List will be empty slice if not exist
	list := in.linkEdgeMap[input.Src]
	for _, id := range list {
		existEdge := in.edges[id]
		if existEdge.Src == input.Src && existEdge.Dst == input.Dst {
			existEdge.UpdateAt = time.Now()
			*input = *existEdge // assign underlying value of pointer
			return nil
		}
	}

	// insert new Edge

	input.ID = mustUniqueEdge(in.edges)
	input.UpdateAt = time.Now()
	eCopy := new(graph.Edge)
	*eCopy = *input
	in.edges[eCopy.ID] = eCopy

	in.linkEdgeMap[input.Src] = append(in.linkEdgeMap[input.Src], eCopy.ID)

	// log.Println("DEBUG inMEMORY LINKGRAPH edges length:", len(in.edges), eCopy.Src)
	return nil
}

// Edges implements graph.Graph.
func (in *InMemory) Edges(fromID uuid.UUID, toID uuid.UUID, update time.Time) (graph.EdgeIterator, error) {
	from, to := fromID.String(), toID.String()

	in.mu.RLock()
	var list []*graph.Edge
	for edgeID, edge := range in.edges {
		if id := edgeID.String(); id >= from && id < to && edge.UpdateAt.Before(update) {
			list = append(list, edge)
		}
	}
	in.mu.RUnlock()
	return &edgeIterator{
		mem:  in,
		list: list,
		idx:  0,
	}, nil
}
