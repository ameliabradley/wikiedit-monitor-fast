package recentchanges

import (
	"log"
	"net/url"

	"github.com/r3labs/sse"
)

const wikiSSEServer = "https://stream.wikimedia.org"
const wikiSSEPath = "/v2/stream/recentchange"

// NewSSE creates a new SSE for listening
func NewSSE(hidebots bool) *sse.Client {
	url, err := url.Parse(wikiSSEServer + wikiSSEPath)
	if err != nil {
		log.Fatal("Could not instantiate stream listener: Parse error")
	}

	q := url.Query()

	if hidebots {
		q.Set("hidebots", "1")
		url.RawQuery = q.Encode()
	}

	fullURL := url.String()
	log.Printf("Creating new client %q", fullURL)
	return sse.NewClient(fullURL)
}
