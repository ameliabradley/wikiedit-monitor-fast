package recentchanges

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

const WikiSSEServer = "https://stream.wikimedia.org"
const WikiSSEPath = "/v2/stream/recentchange"

// Listener listens to recent changes
type Listener interface {
	Listen(lo ListenOptions, handler Handler)
}

type streamListener struct {
	logger *logrus.Logger
	client wiki.SSEClient
}

// ListenOptions are options for what to listen to
type ListenOptions struct {
	Hidebots bool
	Wikis    []string
}

// Handler handles recent changes coming from a stream
type Handler func(rc RecentChange, err error)

// NewStreamListener creates a new stream for listening to wiki changes
func NewStreamListener(client wiki.SSEClient, logger *logrus.Logger) Listener {
	return &streamListener{
		logger: logger,
		client: client,
	}
}

// Listen to the given wikis, with the given handler
func (sl *streamListener) Listen(lo ListenOptions, handler Handler) {
	url := getURLFromOptions(lo)
	go sl.client.Subscribe(url, func(event *sse.Event) {
		rc, err := sl.handleMessage(lo.Wikis, event.Data, handler)
		handler(rc, err)
	})
}

func getURLFromOptions(lo ListenOptions) string {
	url, err := url.Parse(WikiSSEServer + WikiSSEPath)
	if err != nil {
		log.Fatal("Could not instantiate stream listener: Parse error")
	}

	q := url.Query()

	if lo.Hidebots {
		q.Set("hidebots", "1")
		url.RawQuery = q.Encode()
	}

	return url.String()
}

func (sl *streamListener) handleMessage(wikis []string, data []byte, handler Handler) (RecentChange, error) {
	rc := RecentChange{}
	err := json.Unmarshal(data, &rc)

	if err != nil {
		data := string(data[:])
		sl.logger.WithError(err).WithFields(logrus.Fields{
			"data": data,
		}).Error("There was an error decoding")
		return rc, err
	}

	for _, wiki := range wikis {
		if rc.Wiki == wiki {
			return rc, nil
		}
	}

	return RecentChange{}, nil
}
