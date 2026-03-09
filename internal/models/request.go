package models

// AnswerRequest represents an incoming request to the /v1/answer endpoint.
type AnswerRequest struct {
	RequestID          string            `json:"request_id"`
	ParticipantID      string            `json:"participant_id"`
	AgentID            string            `json:"agent_id"`
	SessionID          string            `json:"session_id"`
	Question           string            `json:"question"`
	RequiredCategories []string          `json:"required_categories"`
	Channel            string            `json:"channel"`
	Timestamp          int64             `json:"timestamp"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}
