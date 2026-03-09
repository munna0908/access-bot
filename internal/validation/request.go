package validation

import (
	"errors"
	"strings"

	"github.com/access-bot/internal/domain"
	"github.com/access-bot/internal/models"
)

// ValidationError represents a request validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidateAnswerRequest validates the incoming answer request.
func ValidateAnswerRequest(req *models.AnswerRequest) error {
	var errs []string

	if strings.TrimSpace(req.RequestID) == "" {
		errs = append(errs, "request_id is required")
	}

	if strings.TrimSpace(req.ParticipantID) == "" {
		errs = append(errs, "participant_id is required")
	}

	if strings.TrimSpace(req.AgentID) == "" {
		errs = append(errs, "agent_id is required")
	}

	if strings.TrimSpace(req.SessionID) == "" {
		errs = append(errs, "session_id is required")
	}

	if strings.TrimSpace(req.Question) == "" {
		errs = append(errs, "question is required")
	}

	if len(req.RequiredCategories) == 0 {
		errs = append(errs, "required_categories must contain at least one category")
	}

	// Validate each category
	for _, cat := range req.RequiredCategories {
		if !domain.IsValidCategory(cat) {
			errs = append(errs, "invalid category: "+cat)
		}
	}

	if strings.TrimSpace(req.Channel) == "" {
		errs = append(errs, "channel is required")
	}

	if req.Timestamp <= 0 {
		errs = append(errs, "timestamp must be a positive integer")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// ValidateCategories checks if all categories are valid.
func ValidateCategories(categories []string) ([]string, error) {
	var invalid []string
	var valid []string

	for _, cat := range categories {
		if domain.IsValidCategory(cat) {
			valid = append(valid, cat)
		} else {
			invalid = append(invalid, cat)
		}
	}

	if len(invalid) > 0 {
		return nil, errors.New("invalid categories: " + strings.Join(invalid, ", "))
	}

	return valid, nil
}
