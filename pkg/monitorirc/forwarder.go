package monitorirc

import (
	"encoding/json"
	"fmt"

	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/irc"
)

// Forwarder forwards wikimedia data
type Forwarder interface {
	Forward(lo recentchanges.ListenOptions, subj string)
}

type monitorIrcForwarder struct {
	natsconn *nats.Conn
	listener irc.Listener
	logger   *logrus.Logger
}

// DefaultForwardSubj is the default nats bus subject for incoming sse data
const DefaultForwardSubj = "recentchange.irc"

// NewForwarder creates a new service for forwarding wikimedia sse data to nats
func NewForwarder(listener irc.Listener, natsconn *nats.Conn, logger *logrus.Logger) Forwarder {
	return &monitorIrcForwarder{
		natsconn: natsconn,
		listener: listener,
		logger:   logger,
	}
}

func (f *monitorIrcForwarder) Forward(lo recentchanges.ListenOptions, subj string) {
	f.listener.Listen(lo, func(rc irc.RecentChange, err error) {
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

		f.logger.WithFields(logrus.Fields{
			"rc": fmt.Sprintf("%+v", rc),
		}).Info("Publishing recent change")
		f.natsconn.Publish(subj, data)
	})
}
