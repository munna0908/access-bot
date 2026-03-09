package accesslayer

import (
	"context"
	"strings"
	"time"

	"github.com/access-bot/internal/config"
	"github.com/access-bot/internal/domain"
	"github.com/access-bot/internal/logging"
	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/prompts"
	"github.com/access-bot/internal/providers"
	"github.com/access-bot/internal/validation"
)

// Service implements the access layer business logic.
type Service struct {
	sessionProvider       providers.SessionProvider
	categoryIndexProvider providers.CategoryIndexProvider
	filesystemProvider    providers.FilesystemProvider
	llmProvider           providers.LLMProvider
	logger                *logging.Logger
}

// NewService creates a new access layer service.
func NewService(
	sessionProvider providers.SessionProvider,
	categoryIndexProvider providers.CategoryIndexProvider,
	filesystemProvider providers.FilesystemProvider,
	llmProvider providers.LLMProvider,
	logger *logging.Logger,
) *Service {
	return &Service{
		sessionProvider:       sessionProvider,
		categoryIndexProvider: categoryIndexProvider,
		filesystemProvider:    filesystemProvider,
		llmProvider:           llmProvider,
		logger:                logger,
	}
}

// ProcessAnswerRequest handles the full answer request flow.
func (s *Service) ProcessAnswerRequest(ctx context.Context, req *models.AnswerRequest) *models.AnswerResponse {
	reqLogger := s.logger.WithRequest(req.RequestID, req.ParticipantID, req.SessionID)

	reqLogger.Info(ctx, "processing_request",
		"required_categories", req.RequiredCategories,
		"channel", req.Channel,
	)

	// Step 1: Validate request body
	if err := validation.ValidateAnswerRequest(req); err != nil {
		reqLogger.Error(ctx, "validation_failed", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusInvalidRequest), string(domain.ReasonValidationFailed))
	}

	// Step 2: Load session
	session, err := s.sessionProvider.GetSession(ctx, req.SessionID)
	if err != nil {
		reqLogger.Error(ctx, "session_load_failed", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonSessionNotFound))
	}
	if session == nil {
		reqLogger.Error(ctx, "session_not_found")
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonSessionNotFound))
	}

	// Step 3: Validate session
	if reason := s.validateSession(ctx, reqLogger, session, req); reason != "" {
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusPermissionDen), reason)
	}

	// Step 4: Validate requested categories
	validCategories, err := validation.ValidateCategories(req.RequiredCategories)
	if err != nil {
		reqLogger.Error(ctx, "invalid_categories", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusInvalidRequest), string(domain.ReasonInvalidCategory))
	}

	// Step 5 & 6: Map categories to required scopes and check scope coverage
	missingScopes := s.checkScopeCoverage(validCategories, session.ApprovedScopes)
	if len(missingScopes) > 0 {
		reqLogger.Error(ctx, "insufficient_scopes", "missing_scopes", missingScopes)
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusPermissionDen), string(domain.ReasonInsufficientAuthorizedScope))
	}

	// Step 7: Load category CID index
	categoryIndex, err := s.categoryIndexProvider.GetCategoryIndex(ctx, req.ParticipantID)
	if err != nil {
		reqLogger.Error(ctx, "category_index_load_failed", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonCategoryNotFound))
	}
	if categoryIndex == nil {
		reqLogger.Error(ctx, "category_index_not_found")
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonCategoryNotFound))
	}

	// Step 8: Fetch markdown by CID for each requested category
	categoryContents, fetchedCategories, err := s.fetchCategoryContents(ctx, reqLogger, validCategories, categoryIndex)
	if err != nil {
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonCIDFetchFailed))
	}

	if len(categoryContents) == 0 {
		reqLogger.Error(ctx, "no_category_content_fetched")
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonInsufficientAuthorizedContext))
	}

	// Step 9: Build prompt
	userPrompt := prompts.BuildUserPrompt(req.Question, categoryContents)
	systemPrompt := prompts.GetSystemPrompt()

	reqLogger.Debug(ctx, "prompt_built", "categories_count", len(categoryContents))

	// Step 10: Call Claude
	answer, err := s.llmProvider.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		reqLogger.Error(ctx, "llm_call_failed", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonLLMCallFailed))
	}

	// Step 11: Check for empty response
	answer = strings.TrimSpace(answer)
	if answer == "" {
		reqLogger.Error(ctx, "empty_llm_response")
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonEmptyResponse))
	}

	// Step 12: Return successful answer
	modelName := s.llmProvider.GetModelName()
	response := models.NewSuccessResponse(req.RequestID, answer, modelName, fetchedCategories)

	reqLogger.LogRequestResult(ctx, response.Status, response.UsedCategories, response.Reason)

	return response
}

// validateSession checks all session validation rules.
func (s *Service) validateSession(ctx context.Context, reqLogger *logging.RequestLogger, session *models.Session, req *models.AnswerRequest) string {
	// Check participant_id matches
	if session.ParticipantID != req.ParticipantID {
		reqLogger.Error(ctx, "session_participant_mismatch",
			"session_participant", session.ParticipantID,
			"request_participant", req.ParticipantID,
		)
		return string(domain.ReasonSessionParticipantMismatch)
	}

	// Check agent_id matches
	if session.AgentID != req.AgentID {
		reqLogger.Error(ctx, "session_agent_mismatch",
			"session_agent", session.AgentID,
			"request_agent", req.AgentID,
		)
		return string(domain.ReasonSessionAgentMismatch)
	}

	// Check not expired
	now := time.Now().Unix()
	if session.ExpiresAt <= now {
		reqLogger.Error(ctx, "session_expired",
			"expires_at", session.ExpiresAt,
			"current_time", now,
		)
		return string(domain.ReasonSessionExpired)
	}

	// Check not revoked
	if session.Revoked {
		reqLogger.Error(ctx, "session_revoked")
		return string(domain.ReasonSessionRevoked)
	}

	// Check usage limit
	if session.RemainingUses <= 0 {
		reqLogger.Error(ctx, "usage_limit_exceeded",
			"remaining_uses", session.RemainingUses,
		)
		return string(domain.ReasonUsageLimitExceeded)
	}

	return ""
}

// checkScopeCoverage verifies all required scopes are present.
func (s *Service) checkScopeCoverage(categories []string, approvedScopes []string) []string {
	scopeSet := make(map[string]bool)
	for _, scope := range approvedScopes {
		scopeSet[scope] = true
	}

	var missingScopes []string
	for _, category := range categories {
		requiredScope, ok := config.GetRequiredScope(category)
		if !ok {
			continue // Category validation already happened
		}
		if !scopeSet[requiredScope] {
			missingScopes = append(missingScopes, requiredScope)
		}
	}

	return missingScopes
}

// fetchCategoryContents retrieves markdown content for each category.
func (s *Service) fetchCategoryContents(
	ctx context.Context,
	reqLogger *logging.RequestLogger,
	categories []string,
	categoryIndex *models.CategoryIndex,
) ([]prompts.CategoryContent, []string, error) {
	var contents []prompts.CategoryContent
	var fetchedCategories []string

	for _, category := range categories {
		cid, ok := categoryIndex.Categories[category]
		if !ok {
			reqLogger.Warn(ctx, "category_cid_not_found", "category", category)
			continue
		}

		fileContent, err := s.filesystemProvider.FetchByCID(ctx, cid)
		if err != nil {
			reqLogger.Error(ctx, "cid_fetch_error", "category", category, "cid", cid, "error", err.Error())
			continue
		}
		if fileContent == nil {
			reqLogger.Warn(ctx, "cid_content_not_found", "category", category, "cid", cid)
			continue
		}

		if strings.TrimSpace(fileContent.Content) == "" {
			reqLogger.Warn(ctx, "empty_content", "category", category, "cid", cid)
			continue
		}

		contents = append(contents, prompts.CategoryContent{
			Category: category,
			Content:  fileContent.Content,
		})
		fetchedCategories = append(fetchedCategories, category)
	}

	return contents, fetchedCategories, nil
}

// failureResponse creates a failure response and logs the result.
func (s *Service) failureResponse(ctx context.Context, reqLogger *logging.RequestLogger, requestID, status, reason string) *models.AnswerResponse {
	response := models.NewFailureResponse(requestID, status, reason)
	reqLogger.LogRequestResult(ctx, response.Status, response.UsedCategories, response.Reason)
	return response
}
