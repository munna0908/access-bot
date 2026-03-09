package models

// AnswerResponse represents the response from the /v1/answer endpoint.
type AnswerResponse struct {
	RequestID      string   `json:"request_id"`
	Status         string   `json:"status"`
	Answer         string   `json:"answer"`
	UsedCategories []string `json:"used_categories"`
	LLMModel       *string  `json:"llm_model"`
	Reason         *string  `json:"reason"`
}

// NewSuccessResponse creates a successful answer response.
func NewSuccessResponse(requestID, answer, model string, usedCategories []string) *AnswerResponse {
	return &AnswerResponse{
		RequestID:      requestID,
		Status:         "ok",
		Answer:         answer,
		UsedCategories: usedCategories,
		LLMModel:       &model,
		Reason:         nil,
	}
}

// NewFailureResponse creates a failure response with the given status and reason.
func NewFailureResponse(requestID, status, reason string) *AnswerResponse {
	return &AnswerResponse{
		RequestID:      requestID,
		Status:         status,
		Answer:         "Sorry, I can't answer that.",
		UsedCategories: []string{},
		LLMModel:       nil,
		Reason:         &reason,
	}
}
