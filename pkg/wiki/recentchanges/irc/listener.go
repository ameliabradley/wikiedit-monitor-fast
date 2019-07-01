package irc

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v3"
)

// RecentChange represents a recent change on Wikimedia via the IRC stream
type RecentChange struct {
	Channel    string
	Page       string
	Flags      string
	URL        string
	User       string
	Changesize string
	Comment    string
}

func (rc *RecentChange) Normalize() (recentchanges.NormalizedRecentChange, error) {
	parsedURL, err := url.Parse(rc.URL)
	if err != nil {
		return recentchanges.NormalizedRecentChange{}, err
	}

	query := parsedURL.Query()
	id, err := strconv.Atoi(query.Get("rcid"))
	if err != nil {
		id = -1
	}

	new, err := strconv.Atoi(query.Get("diff"))
	if err != nil {
		new = -1
	}

	old, err := strconv.Atoi(query.Get("oldid"))
	if err != nil {
		old = -1
	}

	parts := strings.Split(parsedURL.Hostname(), ".")
	wiki := parts[0]

	bot := strings.Contains(rc.Flags, "B")
	minor := strings.Contains(rc.Flags, "M")

	rcType := ""
	if strings.Contains(rc.Flags, "N") {
		rcType = "new"
	}

	if new != -1 && old != -1 {
		rcType = "edit"
	}

	return recentchanges.NormalizedRecentChange{
		ID:      id,
		Type:    rcType,
		Title:   rc.Page,
		Comment: rc.Comment,
		User:    rc.User,
		Bot:     bot,
		Wiki:    wiki,
		Minor:   minor,
		Revision: recentchanges.Revision{
			New: new,
			Old: old,
		},
		Source: recentchanges.SourceIRC,
	}, nil
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

// DefaultAddr is the default address to connect to via TCP
const DefaultAddr = "irc.wikimedia.org:6667"

var stripper = regexp.MustCompile(`\x1f|\x02|\x12|\x0f|\x16|\x03(?:\d{1,2}(?:,\d{1,2})?)?`)
var parser = regexp.MustCompile(`PRIVMSG (?P<channel>#[A-Za-z.]+) :\[\[(?P<page>.+)\]\] (?P<flags>.+)? (?P<url>https:\/\/[^ ]+) \* (?P<user>.+) \* (?P<changesize>\(.+\)) ?(?P<comment>.+)?`)

// NewListener creates a new IRC Listener
func NewListener(o Options, logger *logrus.Logger) Listener {
	return &ircListener{
		options: o,
		logger:  logger,
	}
}

func (l *ircListener) Listen(lo recentchanges.ListenOptions, handler Handler) {
	l.logger.WithFields(logrus.Fields{
		"url": DefaultAddr,
	}).Info("Listening")
	conn, err := net.Dial("tcp", DefaultAddr)
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
}

func newListenerHandler(listener *ircListener, lo recentchanges.ListenOptions, handler Handler) irc.Handler {
	return &listenHandler{
		lo:       lo,
		handler:  handler,
		listener: listener,
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
	message := string(stripper.ReplaceAll([]byte(full), []byte{}))

	l.listener.logger.WithFields(logrus.Fields{
		"command": m.Command,
		"data":    message,
	}).Debug("Received data")

	if !c.FromChannel(m) {
		return
	}

	matches := findNamedMatches(parser, message)
	rc := RecentChange{
		Channel:    matches["channel"],
		Page:       matches["page"],
		Flags:      matches["flags"],
		URL:        matches["url"],
		User:       matches["user"],
		Changesize: matches["changesize"],
		Comment:    matches["comment"],
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
