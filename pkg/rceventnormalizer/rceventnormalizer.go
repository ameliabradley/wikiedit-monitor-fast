package rceventnormalizer

import (
	"encoding/json"
	"fmt"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/monitorirc"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/monitorsse"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/irc"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/sse"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type RcEventNormalizer struct {
	logger   *logrus.Logger
	natsconn *nats.Conn
}

const DefaultNormalizedSubj = "recentchanges.normalized"

func NewNormalizer(natsconn *nats.Conn, logger *logrus.Logger) *RcEventNormalizer {
	return &RcEventNormalizer{
		logger:   logger,
		natsconn: natsconn,
	}
}

func (n *RcEventNormalizer) Normalize() {
	n.natsconn.Subscribe(monitorsse.DefaultForwardSubj, func(msg *nats.Msg) {
		n.logger.WithFields(logrus.Fields{
			"data": string(msg.Data),
		}).Debug("Received sse data")

		rc := sse.RecentChange{}
		err := json.Unmarshal(msg.Data, &rc)
		if err != nil {
			n.logger.WithError(err).Error("Could not unmarshal")
			return
		}

		switch rc.Type {
		case "new":
		case "edit":
		default:
			return
		}

		normalized := rc.Normalize()
		n.logger.WithFields(logrus.Fields{
			"msg": fmt.Sprintf("%+v", normalized),
		}).Info("Normalized sse data")

		data, err := json.Marshal(normalized)
		if err != nil {
			n.logger.WithError(err).Error("Could not marshal")
			return
		}

		n.natsconn.Publish(DefaultNormalizedSubj, data)
	})

	n.natsconn.Subscribe(monitorirc.DefaultForwardSubj, func(msg *nats.Msg) {
		n.logger.WithFields(logrus.Fields{
			"data": string(msg.Data),
		}).Debug("Received irc data")

		rc := irc.RecentChange{}
		err := json.Unmarshal(msg.Data, &rc)
		if err != nil {
			n.logger.WithError(err).Error("Could not unmarshal")
			return
		}

		normalized, err := rc.Normalize()
		if err != nil {
			n.logger.WithFields(logrus.Fields{
				"data": fmt.Sprintf("%+v", rc),
			}).WithError(err).Error("Could not normalize irc data")
			return
		}

		n.logger.WithFields(logrus.Fields{
			"msg": fmt.Sprintf("%+v", normalized),
		}).Info("Normalized irc data")

		data, err := json.Marshal(normalized)
		if err != nil {
			n.logger.WithError(err).Error("Could not marshal")
			return
		}

		n.natsconn.Publish(DefaultNormalizedSubj, data)
	})
}
