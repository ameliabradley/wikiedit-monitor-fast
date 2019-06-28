package monitor

import (
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/diffs"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/sirupsen/logrus"
)

type QueueRevision = func(revision int)

// Monitor handles recent changes as they arrive
type Monitor struct {
	logger     *logrus.Logger
	stream     recentchanges.Listener
	diffQueuer diffs.DiffQueuer
	diffParser diffs.DiffParser
	archiver   Archiver
}

// NewMonitor creates a handler for recent changes
func NewMonitor(stream recentchanges.Listener, diffQueuer diffs.DiffQueuer, diffParser diffs.DiffParser, archiver Archiver, logger *logrus.Logger) Monitor {
	return Monitor{
		logger:     logger,
		stream:     stream,
		diffQueuer: diffQueuer,
		diffParser: diffParser,
		archiver:   archiver,
	}
}

func (m Monitor) Start(o recentchanges.ListenOptions) {
	m.stream.Listen(o, m.handleRecentChange)
}

func (m Monitor) handleRecentChange(rc recentchanges.RecentChange) {
	// s, _ := json.MarshalIndent(rc, "", "\t")
	// m.logger.Println(string(s))

	if rc.LogAction == recentchanges.LogActionDelete {
		m.logger.Info("Recent change action delete noted")
	}

	if rc.Revision.New == nil {
		m.logger.Debug("Recent change has no new revision id... discarding")
		return
	}

	new := *rc.Revision.New
	m.diffQueuer.Queue(new, func(queryResult []byte, err error) {
		m.handleFetchResponse(new, queryResult, err)
	})
}

func (m Monitor) handleFetchResponse(revision int, queryResult []byte, err error) {
	if err != nil {
		m.logger.WithError(err).Error("Received diffQueuer error")
		return
	}

	_, err = m.diffParser.Parse(queryResult)
	if err != nil {
		m.logger.WithError(err).WithFields(logrus.Fields{
			"body": string(queryResult),
		}).Error("Encountered parsing error")
	}

	m.archiver.Archive(revision, queryResult)
}
