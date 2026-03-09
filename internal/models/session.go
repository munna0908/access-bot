package models

// Session represents a validated session with approved scopes.
type Session struct {
	SessionID      string   `json:"session_id"`
	ParticipantID  string   `json:"participant_id"`
	AgentID        string   `json:"agent_id"`
	ApprovedScopes []string `json:"approved_scopes"`
	ExpiresAt      int64    `json:"expires_at"`
	RemainingUses  int      `json:"remaining_uses"`
	Revoked        bool     `json:"revoked"`
}

// CategoryIndex represents the CID mapping for a participant's data categories.
type CategoryIndex struct {
	ParticipantID string            `json:"participant_id"`
	Categories    map[string]string `json:"categories"`
}

// FileContent represents content fetched from the filesystem by CID.
type FileContent struct {
	CID         string `json:"cid"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}
