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

// ValidateSessionRequest represents a request to validate a session.
type ValidateSessionRequest struct {
	ParticipantID      string   `json:"participantId"`
	AgentID            string   `json:"agentId"`
	SessionID          string   `json:"sessionId"`
	RequiredCategories []string `json:"requiredCategories"`
	RequiredScopes     []string `json:"requiredScopes"`
	CurrentTime        int64    `json:"currentTime"`
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
