package diffs

import "github.com/sirupsen/logrus"

type DiffQueuer interface {
	Queue(revision int, cb HandleFetchResponse)
}

type fetchRequest struct {
	revid int
	cb    HandleFetchResponse
}

type HandleFetchResponse func([]byte, error)

type DiffQueue struct {
	logger *logrus.Logger
	queue  chan fetchRequest
}

func NewDiffQueuer(logger *logrus.Logger, df DiffFetcher) DiffQueuer {
	dq := DiffQueue{
		logger: logger,
		queue:  make(chan fetchRequest, 100),
	}
	go func() {
		for {
			queue := <-dq.queue
			body, err := df.Fetch(queue.revid)
			queue.cb(body, err)
		}
	}()
	return dq
}

// Queue fetches the revision sequentially
func (mc DiffQueue) Queue(revision int, cb HandleFetchResponse) {
	mc.logger.WithFields(logrus.Fields{
		"total": len(mc.queue),
	}).Info("Queueing revision")
	mc.queue <- fetchRequest{
		revid: revision,
		cb:    cb,
	}
}
