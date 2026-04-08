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

	// Step 2: Validate requested categories
	validCategories, err := validation.ValidateCategories(req.RequiredCategories)
	if err != nil {
		reqLogger.Error(ctx, "invalid_categories", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusInvalidRequest), string(domain.ReasonInvalidCategory))
	}

	// Step 3: Map categories to required scopes
	requiredScopes := s.getRequiredScopes(validCategories)

	// Step 4: Validate session via participant-intelligence-service
	validateReq := &models.ValidateSessionRequest{
		ParticipantID:      req.ParticipantID,
		AgentID:            req.AgentID,
		SessionID:          req.SessionID,
		RequiredCategories: validCategories,
		RequiredScopes:     requiredScopes,
		CurrentTime:        time.Now().Unix(),
	}
	if err := s.sessionProvider.ValidateSession(ctx, validateReq); err != nil {
		reason := s.mapValidationError(err)
		reqLogger.Error(ctx, "session_validation_failed", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusPermissionDen), reason)
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

	// Step 9: Build prompt — use restaurant-aware prompt when context is provided
	var userPrompt, systemPrompt string
	if req.RestaurantContext != "" {
		userPrompt = prompts.BuildRestaurantUserPrompt(req.Question, req.RestaurantContext, categoryContents)
		systemPrompt = prompts.GetRestaurantSystemPrompt()
	} else {
		userPrompt = prompts.BuildUserPrompt(req.Question, categoryContents)
		systemPrompt = prompts.GetSystemPrompt()
	}

	reqLogger.Debug(ctx, "prompt_built", "categories_count", len(categoryContents))

	// Step 10: Call Claude
	answer, err := s.llmProvider.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		reqLogger.Error(ctx, "llm_call_failed", "error", err.Error())
		return s.failureResponse(ctx, reqLogger, req.RequestID, string(domain.StatusCannotAnswer), string(domain.ReasonLLMCallFailed))
	}

	// Step 11: Check for empty response
	answer = strings.TrimSpace(answer)

	// For restaurant requests the response must be JSON — extract it in case
	// the LLM prepended reasoning or explanation before the JSON object.
	if req.RestaurantContext != "" {
		answer = extractLastJSONObject(answer)
	}

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

// getRequiredScopes maps categories to their required scopes.
func (s *Service) getRequiredScopes(categories []string) []string {
	var scopes []string
	for _, category := range categories {
		if scope, ok := config.GetRequiredScope(category); ok {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

// mapValidationError converts validation errors to domain reasons.
func (s *Service) mapValidationError(err error) string {
	errMsg := err.Error()
	switch errMsg {
	case "not_found":
		return string(domain.ReasonSessionNotFound)
	case "expired":
		return string(domain.ReasonSessionExpired)
	case "revoked":
		return string(domain.ReasonSessionRevoked)
	case "exhausted":
		return string(domain.ReasonUsageLimitExceeded)
	case "agent_mismatch":
		return string(domain.ReasonSessionAgentMismatch)
	case "participant_mismatch":
		return string(domain.ReasonSessionParticipantMismatch)
	case "missing_scope":
		return string(domain.ReasonInsufficientAuthorizedScope)
	default:
		return string(domain.ReasonSessionNotFound)
	}
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

// extractLastJSONObject extracts the last JSON object from s.
// Some LLMs prepend reasoning text before the JSON — this strips it out.
// If the response is already pure JSON it is returned unchanged.
func extractLastJSONObject(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "{") {
		return s
	}
	// Find the last line that begins with '{' (JSON always starts on its own line)
	if idx := strings.LastIndex(s, "\n{"); idx != -1 {
		return strings.TrimSpace(s[idx:])
	}
	return s
}

// failureResponse creates a failure response and logs the result.
func (s *Service) failureResponse(ctx context.Context, reqLogger *logging.RequestLogger, requestID, status, reason string) *models.AnswerResponse {
	response := models.NewFailureResponse(requestID, status, reason)
	reqLogger.LogRequestResult(ctx, response.Status, response.UsedCategories, response.Reason)
	return response
}
