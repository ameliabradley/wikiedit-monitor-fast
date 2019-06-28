package recentchanges

import (
	"encoding/json"

	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

// Listener listens to recent changes
type Listener interface {
	Listen(lo ListenOptions, handler Handler)
}

type streamListener struct {
	logger *logrus.Logger
	client SSEClient
}

// ListenOptions are options for what to listen to
type ListenOptions struct {
	Wikis []string
}

// Handler handles recent changes coming from a stream
type Handler func(rc RecentChange)

// SSEClient implements the subscribe method for an SSE client
type SSEClient interface {
	Subscribe(stream string, handler func(msg *sse.Event)) error
}

// NewStreamListener creates a new stream for listening to wiki changes
func NewStreamListener(client SSEClient, logger *logrus.Logger) Listener {
	return &streamListener{
		logger: logger,
		client: client,
	}
}

// Listen to the given wikis, with the given handler
func (sl *streamListener) Listen(lo ListenOptions, handler Handler) {
	go sl.client.Subscribe("messages", func(event *sse.Event) {
		sl.handleMessage(lo.Wikis, event.Data, handler)
	})
}

func (sl *streamListener) handleMessage(wikis []string, data []byte, handler Handler) {
	rc := RecentChange{}
	err := json.Unmarshal(data, &rc)

	if err != nil {
		data := string(data[:])
		sl.logger.WithError(err).WithFields(logrus.Fields{
			"data": data,
		}).Error("There was an error decoding")
		return
	}

	for _, wiki := range wikis {
		if rc.Wiki == wiki {
			handler(rc)
		}
	}
}
