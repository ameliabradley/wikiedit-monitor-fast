package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/monitor"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/diffs"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/sirupsen/logrus"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger := logrus.New()
	logger.Info("Starting monitor")

	httpClient := http.Client{
		Timeout: time.Second * 10,
	}

	diffParser := diffs.NewDiffParser(logger)
	diffFetcher := diffs.NewDiffFetcher(logger, httpClient)
	diffQueuer := diffs.NewDiffQueuer(logger, diffFetcher)

	client := wiki.NewSSEClient()
	streamListener := recentchanges.NewStreamListener(client, logger)
	archiver := monitor.NewFileArchiver(logger, "archive")

	m := monitor.NewMonitor(streamListener, diffQueuer, diffParser, archiver, logger)
	m.Start(recentchanges.ListenOptions{
		Hidebots: true,
		Wikis:    []string{"enwiki"},
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
