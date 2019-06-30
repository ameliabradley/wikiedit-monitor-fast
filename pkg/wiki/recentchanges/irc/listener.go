package irc

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v3"
)

type RecentChange struct {
	Channel string
	Page    string
	Flags   string
	URL     string
	User    string
	Diff    string
	Comment string
}

// Handler handles recent changes coming from a stream
type Handler func(rc RecentChange, err error)

// Listener listens to recent changes
type Listener interface {
	Listen(lo recentchanges.ListenOptions, handler Handler)
}

type ircListener struct {
	options Options
	logger  *logrus.Logger
}

// Options for the IRC listener
type Options struct {
	Nick string
	Pass string
	User string
	Name string
}

// NewIRCListener creates a new IRC Listener
func NewIRCListener(o Options, logger *logrus.Logger) Listener {
	return &ircListener{
		options: o,
		logger:  logger,
	}
}

func (l *ircListener) Listen(lo recentchanges.ListenOptions, handler Handler) {
	url := "irc.wikimedia.org:6667"
	l.logger.WithFields(logrus.Fields{
		"url": url,
	}).Info("Listening")
	conn, err := net.Dial("tcp", url)
	if err != nil {
		log.Fatalln(err)
	}

	listenerHandler := newListenerHandler(l, lo, handler)

	config := irc.ClientConfig{
		Nick:    l.options.Nick,
		Pass:    l.options.Pass,
		User:    l.options.User,
		Name:    l.options.Name,
		Handler: listenerHandler,
	}

	// Create the client
	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

type listenHandler struct {
	lo       recentchanges.ListenOptions
	handler  Handler
	listener *ircListener
	stripper *regexp.Regexp
	parser   *regexp.Regexp
}

func newListenerHandler(listener *ircListener, lo recentchanges.ListenOptions, handler Handler) irc.Handler {
	return &listenHandler{
		lo:       lo,
		handler:  handler,
		listener: listener,
		stripper: regexp.MustCompile(`\x1f|\x02|\x12|\x0f|\x16|\x03(?:\d{1,2}(?:,\d{1,2})?)?`),
		parser:   regexp.MustCompile(`PRIVMSG (?P<channel>#[A-Za-z.]+) :\[\[(?P<page>.+)\]\] (?P<flags>.+)? (?P<url>https:\/\/[^ ]+) \* (?P<user>.+) \* (?P<diff>\(.+\)) ?(?P<comment>.+)?`),
	}
}

func (l *listenHandler) Handle(c *irc.Client, m *irc.Message) {
	// 001 is a welcome event, so we join channels there
	if m.Command == "001" {
		for _, wiki := range l.lo.Wikis {
			l.listener.logger.WithFields(logrus.Fields{
				"channels": l.lo.Wikis,
			}).Info("Joining channels")
			command := fmt.Sprintf("JOIN #%s.wikipedia", wiki)
			c.Write(command)
		}
		return
	}

	full := m.String()
	message := string(l.stripper.ReplaceAll([]byte(full), []byte{}))

	l.listener.logger.WithFields(logrus.Fields{
		"command": m.Command,
		"data":    message,
	}).Debug("Received data")

	if !c.FromChannel(m) {
		return
	}

	matches := findNamedMatches(l.parser, message)
	rc := RecentChange{
		Channel: matches["channel"],
		Page:    matches["page"],
		Flags:   matches["flags"],
		URL:     matches["url"],
		User:    matches["user"],
		Diff:    matches["diff"],
		Comment: matches["comment"],
	}

	if rc.Page != "" {
		if l.lo.Hidebots && strings.Contains(rc.Flags, "B") {
			return
		}

		l.handler(rc, nil)
	}
}

func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}
