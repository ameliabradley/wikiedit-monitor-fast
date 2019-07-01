package sse

import (
	"encoding/json"
	"strings"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki"
	"github.com/leebradley/wikiedit-monitor-fast/pkg/wiki/recentchanges"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

// DefaultURL is the default URL to connect to for wikimedia SSE streams
const DefaultURL = "https://stream.wikimedia.org/v2/stream/recentchange"

// RecentChange represents a recent change on wikimedia via the SSE stream
type RecentChange struct {
	Meta struct {
		//   required:
		// 	- topic
		// 	- uri
		// 	- id
		// 	- dt
		// 	- domain
		Topic string `json:"topic"` // The queue topic name this message belongs to.
		// schema_uri:
		//   description: >
		// 	The URI identifying the jsonschema for this event.  This may be just
		// 	a short uri containing only the name and revision at the end of the
		// 	URI path.  e.g. schema_name/12345 is acceptable.  This field
		// 	is not required.
		//   type: string
		URI       string `json:"uri"`        // The unique URI identifying the event. format: uri
		RequestID string `json:"request_id"` // The unique ID of the request that caused the event.
		ID        string `json:"id"`         // The unique ID of this event; should match the dt field.
		// '^[a-fA-F0-9]{8}(-[a-fA-F0-9]{4}){3}-[a-fA-F0-9]{12}$'
		Dt     string `json:"dt"`     // The time stamp of the event, in ISO8601 format. format: date-time
		Domain string `json:"domain"` // The domain the event pertains to. minLength: 1
	} `json:"meta"`

	ID *int `json:"id"` // ID of the recentchange event (rcid). (CAN BE NULL)

	// Type of recentchange event (rc_type). One of "edit", "new", "log",
	// "categorize", or "external". (See Manual:Recentchanges table#rc_type)
	Type string `json:"type"`

	Title string `json:"title"` // Full page name, from Title::getPrefixedText.

	// ID of relevant namespace of affected page (rc_namespace, page_namespace).
	// This is -1 ("Special") for log events.
	Namespace int `json:"namespace"`

	Comment string `json:"comment"` // (rc_comment)

	ParsedComment string `json:"parsedcomment"` // The rc_comment parsed into simple HTML. Optional

	Timestamp int `json:"timestamp"` // Unix timestamp (derived from rc_timestamp).

	User string `json:"user"` // (rc_user_text)

	Bot bool `json:"bot"` // (rc_bot)

	ServerURL string `json:"server_url"` // $wgCanonicalServer

	ServerName string `json:"server_name"` // $wgServerName

	ServerScriptPath string `json:"server_script_path"` // $wgScriptPath

	Wiki string `json:"wiki"` // wfWikiID ($wgDBprefix, $wgDBname)

	// Edit event related fields
	Minor bool `json:"minor"` // (rc_minor).

	// (rc_patrolled). This property only exists if patrolling is supported
	// for this event (based on $wgUseRCPatrol, $wgUseNPPatrol).
	Patrolled bool `json:"patrolled"`

	// Length of old and new change
	Length struct {
		Old *int `json:"old"` // (rc_old_len)
		New *int `json:"new"` // (rc_new_len)
	} `json:"length"`

	// Old and new revision IDs
	Revision struct {
		New *int `json:"new"` // (rc_last_oldid)
		Old *int `json:"old"` // (rc_this_oldid)
	} `json:"revision"`

	// Log event related fields
	LogID *int `json:"log_id"` // (rc_log_id)

	LogType *string `json:"log_type"` // (rc_log_type)

	LogAction string `json:"log_action"` // (rc_log_action)

	/*
		log_params:
		  description: Property only exists if event has rc_params.
		  type: [array, object, string]
		  additionalProperties: true
	*/

	LogActionComment *string `json:"log_action_comment"`
}

func (rc *RecentChange) Normalize() recentchanges.NormalizedRecentChange {
	id := -1
	if rc.Type == "new" && rc.ID != nil {
		id = *rc.ID
	}

	new := -1
	if rc.Revision.New != nil {
		new = *rc.Revision.New
	}

	old := -1
	if rc.Revision.Old != nil {
		old = *rc.Revision.Old
	}

	wiki := strings.Replace(rc.Wiki, "wiki", "", 1)

	return recentchanges.NormalizedRecentChange{
		ID:      id,
		Type:    rc.Type,
		Title:   rc.Title,
		Comment: rc.Comment,
		User:    rc.User,
		Bot:     rc.Bot,
		Wiki:    wiki,
		Minor:   rc.Minor,
		Revision: recentchanges.Revision{
			New: new,
			Old: old,
		},
		Source: recentchanges.SourceSSE,
	}
}

// LogActionDelete is the constant returned by wikimedia for the delete action
const LogActionDelete = "delete"

// Handler handles recent changes coming from a stream
type Handler func(rc RecentChange, err error)

// Listener listens to recent changes
type Listener interface {
	Listen(lo recentchanges.ListenOptions, handler Handler)
}

type sseListener struct {
	logger *logrus.Logger
	client wiki.SSEClient
}

// NewListener creates a new stream for listening to wiki changes
func NewListener(client wiki.SSEClient, logger *logrus.Logger) Listener {
	return &sseListener{
		logger: logger,
		client: client,
	}
}

// Listen to the given wikis, with the given handler
func (sl *sseListener) Listen(lo recentchanges.ListenOptions, handler Handler) {
	sl.logger.WithField("url", DefaultURL).Info("Subscribing to url")
	go sl.client.Subscribe(DefaultURL, func(event *sse.Event) {
		rc, err := sl.handleMessage(lo.Wikis, event.Data, handler)
		if err != nil {
			handler(rc, err)
		}

		if rc.Bot && lo.Hidebots {
			return
		}

		for _, wiki := range lo.Wikis {
			if rc.Wiki == (wiki + "wiki") {
				handler(rc, nil)
			}
		}
	})
}

func (sl *sseListener) handleMessage(wikis []string, data []byte, handler Handler) (RecentChange, error) {
	rc := RecentChange{}
	err := json.Unmarshal(data, &rc)
	if err != nil {
		data := string(data[:])
		sl.logger.WithError(err).WithFields(logrus.Fields{
			"data": data,
		}).Error("There was an error decoding")
		return rc, err
	}

	return rc, nil
}
