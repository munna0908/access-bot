package domain

// Category represents a supported participant data category.
type Category string

const (
	CategoryHealth  Category = "HEALTH"
	CategoryFood    Category = "FOOD"
	CategoryAddress Category = "ADDRESS"
	CategoryPayment Category = "PAYMENT"
)

// ValidCategories is the set of all valid categories for v1.
var ValidCategories = map[Category]bool{
	CategoryHealth:  true,
	CategoryFood:    true,
	CategoryAddress: true,
	CategoryPayment: true,
}

// IsValidCategory checks if a category string is supported.
func IsValidCategory(cat string) bool {
	return ValidCategories[Category(cat)]
}

// Status represents the response status values.
type Status string

const (
	StatusOK             Status = "ok"
	StatusCannotAnswer   Status = "cannot_answer"
	StatusPermissionDen  Status = "permission_denied"
	StatusInvalidRequest Status = "invalid_request"
	StatusInternalError  Status = "internal_error"
)

// Reason represents machine-readable failure reasons.
type Reason string

const (
	ReasonSessionNotFound               Reason = "session_not_found"
	ReasonSessionExpired                Reason = "session_expired"
	ReasonSessionRevoked                Reason = "session_revoked"
	ReasonSessionParticipantMismatch    Reason = "session_participant_mismatch"
	ReasonSessionAgentMismatch          Reason = "session_agent_mismatch"
	ReasonUsageLimitExceeded            Reason = "usage_limit_exceeded"
	ReasonInsufficientAuthorizedScope   Reason = "insufficient_authorized_scope"
	ReasonInsufficientAuthorizedContext Reason = "insufficient_authorized_context"
	ReasonCategoryNotFound              Reason = "category_not_found"
	ReasonCIDFetchFailed                Reason = "cid_fetch_failed"
	ReasonLLMCallFailed                 Reason = "llm_call_failed"
	ReasonEmptyResponse                 Reason = "empty_response"
	ReasonInvalidCategory               Reason = "invalid_category"
	ReasonValidationFailed              Reason = "validation_failed"
)
