package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/rceventnormalizer"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		natsurl string
	)

	flag.StringVar(&natsurl, "natsurl", nats.DefaultURL, "the url used to connect to nats")
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger := logrus.New()
	logger.Info("Starting rc event normalizer")

	natsconn, err := nats.Connect(natsurl)
	if err != nil {
		logger.WithError(err).Fatal("Could not connect to nats")
	}

	normalizer := rceventnormalizer.NewNormalizer(natsconn, logger)
	normalizer.Normalize()

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
