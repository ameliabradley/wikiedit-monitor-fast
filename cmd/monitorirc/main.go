package main

import (
	"flag"
	"strings"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/monitorirc"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges/irc"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		natsurl  string
		nick     string
		pass     string
		user     string
		name     string
		hidebots bool
		wikis    string
	)

	flag.StringVar(&natsurl, "natsurl", nats.DefaultURL, "the url used to connect to nats")
	flag.StringVar(&nick, "nick", "just_here_for_fun", "the irc nick")
	flag.StringVar(&pass, "pass", "password", "the irc password")
	flag.StringVar(&user, "user", "some user", "the irc user")
	flag.StringVar(&name, "name", "Full Name", "the irc full name")
	flag.BoolVar(&hidebots, "hidebots", true, "Whether to hide / ignore bot edits")
	flag.StringVar(&wikis, "wikis", "en", "A comma-delimited list of wikis to listen to")
	flag.Parse()

	logger := logrus.New()
	logger.Info("Starting wikimedia irc monitor")

	client := irc.NewListener(irc.Options{
		Nick: nick,
		Pass: pass,
		User: user,
		Name: name,
	}, logger)

	natsconn, err := nats.Connect(natsurl)
	if err != nil {
		logger.WithError(err).Fatal("Could not connect to nats")
	}

	lo := recentchanges.ListenOptions{
		Hidebots: hidebots,
		Wikis:    strings.Split(wikis, ","),
	}

	forward := monitorirc.NewForwarder(client, natsconn, logger)
	forward.Forward(lo, monitorirc.DefaultForwardSubj)
}
