package recentchanges

import (
	"encoding/json"
	"net/url"

	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

type ListenOptions struct {
	Wikis    []string
	Hidebots bool
}

type StreamListener struct {
	logger *logrus.Logger
}

type MessageHandler func(rc RecentChange)

const streamEndpoint = "https://stream.wikimedia.org/v2/stream/recentchange"

func NewStreamListener(logger *logrus.Logger) StreamListener {
	return StreamListener{
		logger: logger,
	}
}

func (sl *StreamListener) Listen(lo ListenOptions, handler MessageHandler) {
	url, err := url.Parse(streamEndpoint)
	if err != nil {
		sl.logger.WithError(err).Fatal("Could not instantiate stream listener")
	}

	q := url.Query()

	if lo.Hidebots {
		q.Set("hidebots", "1")
		url.RawQuery = q.Encode()
	}

	client := sse.NewClient(url.String())

	go client.Subscribe("messages", func(event *sse.Event) {
		sl.handleMessage(lo.Wikis, event.Data, handler)
	})
}

func (sl *StreamListener) handleMessage(wikis []string, data []byte, handler MessageHandler) {
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
