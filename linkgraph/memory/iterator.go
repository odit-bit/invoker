package memory

import (
	"github.com/odit-bit/invoker/linkgraph/graph"
)

// ===================Link iterator

var _ graph.LinkIterator = (*LinkIterator)(nil)

type LinkIterator struct {
	s    *InMemory
	list []*graph.Link
	idx  int
}

// Close implements graph.Iterator.
func (it *LinkIterator) Close() error {
	it.s = nil
	it.list = it.list[:0]
	it.idx = 0
	return nil
}

// Error implements graph.Iterator.
func (it *LinkIterator) Error() error {
	return nil
}

// Next implements graph.Iterator.
func (it *LinkIterator) Next() bool {
	return it.idx < len(it.list)

}

// Link implements graph.LinkIterator.
func (li *LinkIterator) Link() *graph.Link {
	li.s.mu.RLock()

	val := new(graph.Link)
	l := li.list[li.idx]
	*val = *l
	li.idx++

	li.s.mu.RUnlock()
	return val
}

//========== edge iterator

var _ graph.EdgeIterator = (*edgeIterator)(nil)

type edgeIterator struct {
	mem  *InMemory
	list []*graph.Edge
	idx  int
}

// Close implements graph.EdgeIterator.
func (it *edgeIterator) Close() error {
	it.mem = nil
	it.list = it.list[:0]
	it.idx = 0
	return nil
}

// Edge implements graph.EdgeIterator.
func (it *edgeIterator) Edge() *graph.Edge {
	it.mem.mu.RLock()

	val := new(graph.Edge)
	l := it.list[it.idx]
	*val = *l
	it.idx++

	it.mem.mu.RUnlock()
	return val
}

// Error implements graph.EdgeIterator.
func (*edgeIterator) Error() error {
	return nil //fmt.Errorf("unexpected error ")
}

// Next implements graph.EdgeIterator.
func (it *edgeIterator) Next() bool {
	return it.idx < len(it.list)
}
