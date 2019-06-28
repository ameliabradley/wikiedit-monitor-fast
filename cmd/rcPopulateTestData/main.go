package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"

	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		max int
	)
	flag.IntVar(&max, "max", 0, "The number of samples to gather")
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger := logrus.New()
	logger.WithFields(logrus.Fields{
		"max": max,
	}).Info("Starting rcPopuluateData")

	folder := "../../pkg/wiki/recentchanges/testdata"

	client := wiki.NewSSEClient()
	logger.Info("Subscribing to messages")
	fullURL := recentchanges.WikiSSEServer + recentchanges.WikiSSEPath

	done := make(chan struct{})

	gathered := 0
	var mux sync.Mutex
	go client.Subscribe(fullURL, func(msg *sse.Event) {
		logger.WithFields(logrus.Fields{
			"data": string(msg.Data),
		}).Info("Received data")

		if len(msg.Data) == 0 {
			logger.Info("Empty data: Discarding")
			return
		}

		ts := time.Now().UnixNano()

		path := folder + "/" + strconv.FormatInt(ts, 10) + ".json"
		if fileExists(path) {
			logger.WithFields(logrus.Fields{
				"file": path,
			}).Error("Skipping archiving: File exists")
			return
		}

		logger.WithFields(logrus.Fields{
			"file": path,
		}).Info("Archiving stream as test data")
		err := ioutil.WriteFile(path, msg.Data, 0644)
		if err != nil {
			logger.WithError(err).Error("Could not write file")
		}

		mux.Lock()
		defer mux.Unlock()
		gathered++
		if gathered >= max {
			logger.WithFields(logrus.Fields{
				"max": max,
			}).Info("Gathered maximum stream items")
			done <- struct{}{}
		}
	})

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

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
