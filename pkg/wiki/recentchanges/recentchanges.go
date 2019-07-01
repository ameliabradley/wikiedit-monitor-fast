package recentchanges

// ListenOptions are options for what to listen to
type ListenOptions struct {
	Hidebots bool
	Wikis    []string
}

// NormalizedRecentChange represents a normalization of the recent change streams
type NormalizedRecentChange struct {
	ID int `json:"id"` // ID of the recentchange event (rcid). (-1 is empty) (Must be -1 if type="new")

	// Type of recentchange event (rc_type). One of "edit" or "new"
	Type string `json:"type"`

	Title string `json:"title"` // Full page name, from Title::getPrefixedText.

	Comment string `json:"comment"` // (rc_comment)

	User string `json:"user"` // (rc_user_text)

	Bot bool `json:"bot"` // (rc_bot)

	Wiki string `json:"wiki"` // wfWikiID ($wgDBprefix, $wgDBname)

	// Edit event related fields
	Minor bool `json:"minor"` // (rc_minor).

	// Old and new revision IDs
	Revision Revision `json:"revision"`

	Source string // "irc" or "sse"
}

// Revision represents a Wikimedia revision
type Revision struct {
	New int `json:"new"` // (rc_last_oldid) (-1 is empty)
	Old int `json:"old"` // (rc_this_oldid) (-1 is empty)
}

const (
	// SourceSSE is the NormalizedRecentChange source for SSE stream
	SourceSSE = "sse"

	// SourceIRC is the NormalizedRecentChange source for IRC stream
	SourceIRC = "irc"
)
