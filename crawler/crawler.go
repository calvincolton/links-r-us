package crawler

import (
	"context"
	"net/http"
	"time"

	"github.com/calvincolton/links-r-us/linkgraph/graph"
	"github.com/calvincolton/links-r-us/pipeline"
	index "github.com/calvincolton/links-r-us/textindexer"
	"github.com/google/uuid"
)

// URLGetter is implemented by objects that can perform HTTP GET requests
type URLGetter interface {
	Get(url string) (*http.Response, error)
}

// PrivateNetworkDetector is implemented by objects that can detect whether a host resolves to a private network
type PrivateNetworkDetector interface {
	IsPrivate(host string) (bool, error)
}

// Graph is implemented by objects that can upsert links and edges into the graph instance
type Graph interface {
	UpsertLink(link *graph.Link) error
	UpsertEdge(edge *graph.Edge) error
	RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error
}

// Indexer is implemented by objects that can index the contents of web pages retrieved by the crawler pipeline
type Indexer interface {
	Index(doc *index.Document) error
}

// Config encapsulates the configuration options for creating a new Crawler
type Config struct {
	// A PrivateNetworkDetector instance
	PrivateNetworkDetector PrivateNetworkDetector

	// A URLGetter instance for fetching links.
	URLGetter URLGetter

	// A GraphUpdater instance for addding new links to the link graph.
	Graph Graph

	// A TextIndexer instance for indexing the content of each retrieved link.
	Indexer Indexer

	// The number of concurrent workers used for retrieving links.
	FetchWorkers int
}

// Crawler implements a web page crawling pipeline consisting of the following stages:
// - Given a URL, retrieve the web page contents from the remote server
// - Extract and resolve absolute and relative links from the retrieved page
// - Extract page title and text content from the retrieved page
// - Update the link graph: add new links and create edges between the crawled page and the links within it
// - Index crawled page title and text content
type Crawler struct {
	p *pipeline.Pipeline
}

// NewCrawler returns a new crawler instance
func NewCrawler(cfg Config) *Crawler {
	return &Crawler{
		p: assembleCrawlerPipeline(cfg),
	}
}

// assembleCrawlerPipeline creates the various stages of a crawler pipeline using the options in cfg and assembles them into a pipeline instance
func assembleCrawlerPipeline(cfg Config) *pipeline.Pipeline {
	return pipeline.New(
		pipeline.FixedWorkerPool(
			newLinkFetcher(cfg.URLGetter, cfg.PrivateNetworkDetector),
			cfg.FetchWorkers,
		),
		pipeline.FIFO(newLinkExtractor(cfg.PrivateNetworkDetector)),
		pipeline.FIFO(newTextExtractor()),
		pipeline.Broadcast(
			newGraphUpdater(cfg.Graph),
			newTextIndexer(cfg.Indexer),
		),
	)
}

// Crawl iterates linkIt and sends each link through the crawler pipeline returning the total count of links that went through the pipeline
// Calls to Crawl block until the link iterator is exhausted, an error occurs, or the context is cancelled
func (c *Crawler) Crawl(ctx context.Context, linkIt graph.LinkIterator) (int, error) {
	sink := new(countingSink)
	err := c.p.Process(ctx, &linkSource{linkIt: linkIt}, sink)
	return sink.getCount(), err
}

type linkSource struct {
	linkIt graph.LinkIterator
}

func (ls *linkSource) Error() error              { return ls.linkIt.Error() }
func (ls *linkSource) Next(context.Context) bool { return ls.linkIt.Next() }
func (ls *linkSource) Payload() pipeline.Payload {
	link := ls.linkIt.Link()
	p := payloadPool.Get().(*crawlerPayload)

	p.LinkID = link.ID
	p.URL = link.URL
	p.RetrievedAt = link.RetrievedAt
	return p
}

type countingSink struct {
	count int
}

func (s *countingSink) Consume(_ context.Context, p pipeline.Payload) error {
	s.count++
	return nil
}

func (s *countingSink) getCount() int {
	// The broadcast split-stage sends out two payloads for each incoming link, so we need to divide the total count by 2
	return s.count / 2
}
