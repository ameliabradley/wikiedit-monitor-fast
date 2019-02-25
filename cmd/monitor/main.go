package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/monitor"
	"github.com/r3labs/sse"
)

var addr = flag.String("addr", "stream.wikimedia.org", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger := log.New(os.Stdout, "", log.LstdFlags)
	fetcher := monitor.NewRevisionFetcher(logger)
	handler := monitor.NewRecentChangeHandler(fetcher, logger)
	parser := monitor.CreateMessageParser(handler.Handle, logger)
	url := "https://stream.wikimedia.org/v2/stream/recentchange?hidebots=1"
	client := sse.NewClient(url)
	go client.Subscribe("messages", parser)
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
