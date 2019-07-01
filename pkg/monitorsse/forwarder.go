package monitorsse

import (
	"encoding/json"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/sse"
)

// Forwarder forwards wikimedia data
type Forwarder interface {
	Forward(lo recentchanges.ListenOptions, subj string)
}

type monitorSseForwarder struct {
	natsconn *nats.Conn
	listener sse.Listener
	logger   *logrus.Logger
}

// DefaultForwardSubj is the default nats bus subject for incoming sse data
const DefaultForwardSubj = "recentchange.sse"

// NewForwarder creates a new service for forwarding wikimedia sse data to nats
func NewForwarder(natsconn *nats.Conn, logger *logrus.Logger) Forwarder {
	sseclient := wiki.NewSSEClient()
	return &monitorSseForwarder{
		natsconn: natsconn,
		listener: sse.NewListener(sseclient, logger),
		logger:   logger,
	}
}

func (f *monitorSseForwarder) Forward(lo recentchanges.ListenOptions, subj string) {
	f.listener.Listen(lo, func(rc sse.RecentChange, err error) {
		f.logger.WithFields(logrus.Fields{
			"comment": rc,
		}).Info("Publishing recent change")

		if err != nil {
			f.logger.WithError(err).Error("Encountered error in stream")
			return
		}

		data, err := json.Marshal(rc)
		if err != nil {
			f.logger.WithFields(logrus.Fields{
				"rc": rc,
			}).WithError(err).Error("Could not marshal")
			return
		}

		f.natsconn.Publish(subj, data)
	})
}
