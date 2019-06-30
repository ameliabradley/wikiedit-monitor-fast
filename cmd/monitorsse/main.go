package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/sse"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		natsurl string
	)

	flag.StringVar(&natsurl, "natsurl", nats.DefaultURL, "the url used to connect to nats")
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger := logrus.New()
	logger.Info("Starting wikimedia sse monitor")

	natsclient, _ := nats.Connect(natsurl)

	sseclient := wiki.NewSSEClient()
	streamListener := sse.NewStreamListener(sseclient, logger)
	streamListener.Listen(recentchanges.ListenOptions{
		Hidebots: true,
		Wikis:    []string{"enwiki"},
	}, func(rc sse.RecentChange, err error) {
		data, err := json.Marshal(rc)
		if err != nil {
			logger.WithError(err).Error("Encountered error in stream")
			return
		}

		logger.WithFields(logrus.Fields{
			"comment": rc.Comment,
		}).Info("Publishing recent change")
		natsclient.Publish("recentchange.sse", data)
	})

	done := make(chan struct{})

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
