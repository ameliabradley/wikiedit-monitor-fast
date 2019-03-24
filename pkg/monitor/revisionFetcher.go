package monitor

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type RevisionFetcher interface {
	Queue(revision int, cb HandleFetchResponse)
}

type SuccessCallback = func(data []byte)

type fetchRequest struct {
	revid int
	cb    HandleFetchResponse
}

type HandleFetchResponse func([]byte, error)

type RevisionFetch struct {
	client http.Client
	logger *log.Logger
	queue  chan fetchRequest
}

const baseUrl = "https://en.wikipedia.org/w/api.php?action=compare&format=json&fromrev=%d&torelative=prev"

func NewRevisionFetcher(logger *log.Logger) RevisionFetch {
	mc := RevisionFetch{
		logger: logger,
		client: http.Client{
			Timeout: time.Second * 10,
		},
		queue: make(chan fetchRequest, 100),
	}
	go func() {
		for {
			queue := <-mc.queue
			body, err := mc.fetch(queue.revid)
			queue.cb(body, err)
		}
	}()
	return mc
}

func (mc RevisionFetch) Queue(revision int, cb HandleFetchResponse) {
	mc.logger.Printf("queueing revision [%d queued]\n", len(mc.queue))
	mc.queue <- fetchRequest{
		revid: revision,
		cb:    cb,
	}
}

func (mc RevisionFetch) fetch(revision int) ([]byte, error) {
	url := fmt.Sprintf(baseUrl, revision)
	mc.logger.Printf("Fetching revision: %s\n", url)
	start := time.Now()
	resp, err := mc.client.Get(url)
	if err != nil {
		fmt.Println("error querying: %+v", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading body: %+v", err)
		return nil, err
	}
	diff := time.Now().Sub(start)
	mc.logger.Printf("Revision fetched in %s\n", diff.String())

	// fmt.Println(string(body))
	return body, nil
}
