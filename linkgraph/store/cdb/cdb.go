package cdb

import "database/sql"

var (
	upsertLinkQuery = `
		INSERT INTO links (url, retrieved_at) VALUES ($1, $2) 
		ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2) 
		RETURNING id, retrieved_at
	`

	findLinkQuery = `
		SELECT url, retrieved_at FROM links WHERE id = $1
	`

	linksInPartitionQuery = `
		SELECT id, url, retrieved_at FROM links WHERE id >= $1 AND id < $2 AND retrieved_at < $3
	`

	upsertEdgeQuery = `
		INSERT INTO edges (src, dst, updated_at) VALUES ($1, $2, NOW()) 
		ON CONFLICT (src,dst) DO UPDATE SET updated_at=NOW() 
		RETURNING id, updated_at
	`

	edgesInPartitionQuery = `
		SELECT id, src, dst, updated_at FROM edges WHERE src >= $1 AND src < $2 AND updated_at < $3
	`

	removeStaleEdgesQuery = `
		DELETE FROM edges WHERE src=$1 AND updated_at < $2
	`

	_ graph.Graph = (*CockRoachDBGraph)(nil)
)

// CockroachDBGraph implements a graph that persists its links and edges to a cockroachdb instance.
type CockRoachDBGraph struct {
	db *sql.DB
}

// NewCockroachDbGraph returns a CockroachDbGraph instance that connects to the cockroachdb instance specified by dsn.
func NewCockRoachDBGraph(dsn string) (*CockRoachDBGraph, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &CockRoachDBGraph{db: db}, nil
}

// Close terminates the connection to the backing cockroachdb instance
func (c *CockRoachDBGraph) Close() error {
	return c.db.Close()
}

// UpsertLink creates a new link or updates an existing link
func (c *CockRoachDBGraph) UpsertLink(link *graph.Link) error {
	row := c.db.QueryRow(upsertLinkQuery, link.URL, link.RetrievedAt.UTC())

}
