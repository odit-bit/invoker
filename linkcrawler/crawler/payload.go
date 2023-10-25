package crawler

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/odit-bit/pipeline"
)

// payload implemet pipeline.Payload
// so that can use by pipeline.Processor
type payload struct {
	// populated by input source
	LinkID      uuid.UUID
	URL         string
	RetrievedAt time.Time

	NoFollowLinks []string
	RawContent    bytes.Buffer

	Links       []string
	Title       []byte
	TextContent []byte
}

// Clone implements pipeline.Payload.
func (p *payload) Clone() pipeline.Payload {
	cloneP := payloadPool.Get().(*payload)
	cloneP.LinkID = p.LinkID
	cloneP.URL = p.URL
	cloneP.RetrievedAt = p.RetrievedAt
	cloneP.NoFollowLinks = append([]string(nil), p.NoFollowLinks...)
	cloneP.Links = append([]string(nil), p.Links...)
	cloneP.Title = p.Title
	cloneP.TextContent = p.TextContent

	_, err := io.Copy(&cloneP.RawContent, &p.RawContent)
	if err != nil {
		panic(fmt.Sprintf("[BUG] error cloning payload raw content: %v", err))
	}
	return cloneP
}

// MarkAsProcessed implements pipeline.Payload.
func (p *payload) MarkAsProcessed() {
	p.URL = p.URL[:0]
	p.Links = p.Links[:0]
	p.NoFollowLinks = p.NoFollowLinks[:0]
	p.Title = p.Title[:0]
	p.TextContent = p.TextContent[:0]
	p.RawContent.Reset()
	payloadPool.Put(p)
}
