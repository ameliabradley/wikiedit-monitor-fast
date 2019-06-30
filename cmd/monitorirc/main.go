package main

import (
	"fmt"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/irc"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Info("Starting wikimedia irc monitor")

	client := irc.NewIRCListener(irc.Options{
		Nick: "just_here_for_fun",
		Pass: "password",
		User: "some user",
		Name: "Full Name",
	}, logger)

	client.Listen(recentchanges.ListenOptions{
		Wikis:    []string{"en"},
		Hidebots: true,
	}, func(rc irc.RecentChange, err error) {
		// jsonData, err := json.Marshal(recentChange)
		// if err != nil {
		// 	l.logger.WithFields(logrus.Fields{
		// 		"rc": recentChange,
		// 	}).Error("Could not marhsal")
		// 	return
		// }

		logger.WithFields(logrus.Fields{
			"rc": fmt.Sprintf("%+v", rc),
		}).Info("Received change")
	})
}
