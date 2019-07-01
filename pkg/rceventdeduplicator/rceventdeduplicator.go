package rceventdeduplicator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/rceventnormalizer"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type RcEventDeduplicator struct {
	logger   *logrus.Logger
	natsconn *nats.Conn
	mux      sync.Mutex
	store    map[string]bool
}

const DefaultDeduplicatedSubj = "recentchanges.dedup"

func NewDeduplicator(natsconn *nats.Conn, logger *logrus.Logger) *RcEventDeduplicator {
	return &RcEventDeduplicator{
		logger:   logger,
		natsconn: natsconn,
		store:    make(map[string]bool),
	}
}

func (n *RcEventDeduplicator) checkAndStore(id string) bool {
	n.mux.Lock()
	defer n.mux.Unlock()

	if !n.store[id] {
		n.store[id] = true
		go func() {
			time.Sleep(10 * time.Minute)
			n.logger.WithField("id", id).Info("Discarding old id")
			delete(n.store, id)
		}()
		return false
	}

	return true
}

func (n *RcEventDeduplicator) Deduplicate() {
	n.natsconn.Subscribe(rceventnormalizer.DefaultNormalizedSubj, func(msg *nats.Msg) {
		n.logger.WithFields(logrus.Fields{
			"data": string(msg.Data),
		}).Debug("Received sse data")

		rc := recentchanges.NormalizedRecentChange{}
		err := json.Unmarshal(msg.Data, &rc)
		if err != nil {
			n.logger.WithError(err).Error("Could not unmarshal")
			return
		}

		id := strconv.Itoa(rc.ID) + ":" + strconv.Itoa(rc.Revision.New)
		exists := n.checkAndStore(id)
		if exists {
			n.logger.WithField("id", id).Info("discarding duplicate")
			return
		}

		data, err := json.Marshal(rc)
		if err != nil {
			n.logger.WithError(err).Error("Could not marshal")
			return
		}

		n.logger.WithFields(logrus.Fields{
			"msg": fmt.Sprintf("%+v", rc),
		}).Info("Deduplicated data")

		n.natsconn.Publish(DefaultDeduplicatedSubj, data)
	})
}
