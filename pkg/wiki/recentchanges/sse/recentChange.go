package sse

// RecentChange represents a recent change on wikipedia
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

// LogActionDelete is the constant returned by wikipedia for the delete action
const LogActionDelete = "delete"
