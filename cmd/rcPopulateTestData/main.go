package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger := logrus.New()
	logger.Info("Starting rcPopuluateData")

	// folder := "../../pkg/recentchanges/testdata"

	client := recentchanges.NewSSE(true)
	logger.Info("Subscribing to messages")
	go client.Subscribe("messages", func(msg *sse.Event) {
		logger.WithFields(logrus.Fields{
			"data": msg.Data,
		}).Info(msg.Data)

		// ts := time.Now().Unix()

		// path := folder + "/" + strconv.FormatInt(ts, 10)
		// logger.WithFields(logrus.Fields{
		// 	"file": path,
		// }).Info("Archiving stream as test data")
		// err := ioutil.WriteFile(path, msg.Data, 0644)
		// if err != nil {
		// 	logger.WithError(err).Error("Could not write file")
		// }
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
