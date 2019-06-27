package diffs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type DiffFetch struct {
	client http.Client
	logger *logrus.Logger
}

type DiffFetcher interface {
	Fetch(revision int) ([]byte, error)
}

const baseUrl = "https://en.wikipedia.org/w/api.php?action=compare&format=json&fromrev=%d&torelative=prev"

func NewDiffFetcher(logger *logrus.Logger, client http.Client) DiffFetcher {
	mc := DiffFetch{
		logger: logger,
		client: client,
	}
	return mc
}

func (mc DiffFetch) Fetch(revision int) ([]byte, error) {
	url := fmt.Sprintf(baseUrl, revision)
	mc.logger.WithFields(logrus.Fields{
		"url": url,
	}).Info("Fetching revision")
	start := time.Now()
	resp, err := mc.client.Get(url)
	if err != nil {
		mc.logger.WithError(err).Error("Error querying")
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mc.logger.WithError(err).Error("Error reading body")
		return nil, err
	}
	diff := time.Now().Sub(start)
	mc.logger.WithFields(logrus.Fields{
		"time":  diff.String(),
		"bytes": len(body),
	}).Info("Revision fetched")

	return body, nil
}
