package postgredb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/odit-bit/invoker/linkgraph/graph"
)

const lookupLinkQuery = `
	SELECT id, url, retrieved_at
	FROM links
	WHERE id = $1
`

// LookupLink implements graph.Graph.
func (p *postgre) LookupLink(id uuid.UUID) (*graph.Link, error) {
	var link graph.Link

	err := p.db.QueryRowxContext(context.TODO(), lookupLinkQuery, id).Scan(&link.ID, &link.URL, &link.RetrievedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, graph.ErrNotFound
		}
		return nil, fmt.Errorf("lookup link: %v", err)
	}

	return &link, nil
}
