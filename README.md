# Access Layer Service (v1)

A participant-centric access layer service that validates session permissions via the participant-intelligence-service, fetches authorized participant data, and uses LLMs to answer questions within the approved context.

## Overview

The Access Layer Service acts as a policy boundary between requesting agents (like OpenClaw WhatsApp bot) and participant data. It:

1. Validates session access by calling the participant-intelligence-service
2. Fetches markdown files by CID for approved categories
3. Sends the user question plus approved context to an LLM (Claude or Gemini)
4. Returns the LLM's answer or a safe failure response

**Key principle**: The access layer enforces all permission checks. The LLM is only a reasoning engine—it is never trusted for permission decisions.

## Features

- **Session Validation**: Delegates validation to participant-intelligence-service (`/v1/sessions/validate`)
- **Scope-Based Permissions**: Maps data categories to OAuth-style scopes
- **Content Retrieval**: Fetches markdown content from Pinata IPFS or mock filesystem
- **LLM Integration**: Supports Claude and Gemini APIs
- **Safe Failures**: Returns "Sorry, I can't answer that." for all error conditions

## Request/Response Contract

### POST /v1/answer

**Request:**
```json
{
  "request_id": "req_123",
  "participant_id": "user_001",
  "agent_id": "openclaw_whatsapp_bot",
  "session_id": "sess_123",
  "question": "Order some dinner",
  "required_categories": ["FOOD", "HEALTH"],
  "channel": "whatsapp",
  "timestamp": 1773162200,
  "metadata": {
    "conversation_id": "conv_001",
    "message_id": "msg_001"
  }
}
```

**Success Response:**
```json
{
  "request_id": "req_123",
  "status": "ok",
  "answer": "A vegetarian South Indian dinner would suit your preferences tonight. Avoid dairy-based dishes.",
  "used_categories": ["FOOD", "HEALTH"],
  "llm_model": "claude-sonnet-4-20250514",
  "reason": null
}
```

**Failure Response:**
```json
{
  "request_id": "req_123",
  "status": "permission_denied",
  "answer": "Sorry, I can't answer that.",
  "used_categories": [],
  "llm_model": null,
  "reason": "session_not_found"
}
```

### Status Values
- `ok` - Request processed successfully
- `cannot_answer` - Unable to answer safely (missing data, LLM failure)
- `permission_denied` - Session validation failed (not found, expired, revoked, missing scopes)
- `invalid_request` - Request validation failed
- `internal_error` - Unexpected server error

### GET /health

Returns service health status.

## Category Permission Model

The service supports four data categories, each requiring a specific scope:

| Category | Required Scope |
|----------|---------------|
| HEALTH | health.read |
| FOOD | preferences.food.read |
| ADDRESS | profile.address.read |
| SCHEDULE | schedule.read |

A session must have all required scopes to access the requested categories.

## Session Validation

Session validation is delegated to the **participant-intelligence-service** via `POST /v1/sessions/validate`.

**Validation Request:**
```json
{
  "participantId": "user_001",
  "agentId": "openclaw_whatsapp_bot",
  "sessionId": "sess_123",
  "requiredCategories": ["FOOD", "HEALTH"],
  "requiredScopes": ["preferences.food.read", "health.read"],
  "currentTime": 1773162200
}
```

**Validation Response:**
```json
{
  "valid": true,
  "reason": null
}
```

The intelligence service validates:
1. Session exists
2. Session belongs to the requested participant_id
3. Session is bound to the requesting agent_id
4. Session is not expired
5. Session is not revoked
6. Usage limit is not exceeded
7. All required categories/scopes are authorized

## Running Locally

### Prerequisites
- Go 1.22+
- Anthropic API key or Gemini API key (optional for mock mode)

### Setup

1. Clone the repository:
```bash
cd /path/to/access-bot
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Configure `.env`:
```bash
# LLM Provider (claude, gemini, mock, or auto)
LLM_PROVIDER=auto
ANTHROPIC_API_KEY=sk-ant-api03-xxxxx
# Or use Gemini:
# GEMINI_API_KEY=xxxxx

# Session Provider (intelligence, mock, or auto)
SESSION_PROVIDER=auto
INTELLIGENCE_SERVICE_URL=http://localhost:3000

# Filesystem Provider (pinata, mock, or auto)
FILESYSTEM_PROVIDER=auto
PINATA_GATEWAY_URL=https://your-gateway.mypinata.cloud
```

4. Run the server:
```bash
go run cmd/main.go
```

5. Server starts at `http://localhost:8080`

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | HTTP server port |
| HOST | localhost | HTTP server host |
| LOG_LEVEL | info | Log level (debug, info, warn, error) |
| **LLM Configuration** | | |
| LLM_PROVIDER | auto | LLM provider: claude, gemini, mock, or auto |
| ANTHROPIC_API_KEY | (empty) | Claude API key |
| CLAUDE_MODEL | claude-sonnet-4-20250514 | Claude model to use |
| GEMINI_API_KEY | (empty) | Gemini API key |
| GEMINI_MODEL | gemini-2.0-flash | Gemini model to use |
| **Session Configuration** | | |
| SESSION_PROVIDER | auto | Session provider: intelligence, mock, or auto |
| INTELLIGENCE_SERVICE_URL | (empty) | Base URL of participant-intelligence-service |
| **Filesystem Configuration** | | |
| FILESYSTEM_PROVIDER | auto | Filesystem provider: pinata, mock, or auto |
| PINATA_GATEWAY_URL | (empty) | Pinata gateway URL |
| PINATA_GATEWAY_KEY | (empty) | Pinata gateway key (for dedicated gateways) |
| PINATA_JWT | (empty) | Pinata JWT (for private files) |
| PINATA_PRIVATE | false | Whether files are private |
| **Timeouts** | | |
| HTTP_CLIENT_TIMEOUT | 30 | HTTP client timeout in seconds |
| SERVER_READ_TIMEOUT | 10 | Server read timeout in seconds |
| SERVER_WRITE_TIMEOUT | 30 | Server write timeout in seconds |

### Auto-Detection

When providers are set to `auto`:
- **LLM**: Uses Claude if `ANTHROPIC_API_KEY` is set, else Gemini if `GEMINI_API_KEY` is set, else mock
- **Session**: Uses intelligence service if `INTELLIGENCE_SERVICE_URL` is set, else mock
- **Filesystem**: Uses Pinata if `PINATA_GATEWAY_URL` or `PINATA_JWT` is set, else mock

## Sample Curl Request

```bash
curl -X POST http://localhost:8080/v1/answer \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req_123",
    "participant_id": "user_001",
    "agent_id": "openclaw_whatsapp_bot",
    "session_id": "sess_123",
    "question": "Order some dinner for tonight",
    "required_categories": ["FOOD", "HEALTH"],
    "channel": "whatsapp",
    "timestamp": 1773162200,
    "metadata": {
      "conversation_id": "conv_001",
      "message_id": "msg_001"
    }
  }'
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Mock Providers

For local development, the service uses in-memory mocks when real providers aren't configured.

### Sample Mock Data

**Sessions (sess_123)**:
- participant_id: user_001
- agent_id: openclaw_whatsapp_bot
- approved_scopes: preferences.food.read, health.read
- expires_at: far future
- remaining_uses: 5

**User Categories (user_001)**:
- HEALTH: bafy_health_cid_001
- FOOD: bafy_food_cid_001
- ADDRESS: bafy_address_cid_001
- SCHEDULE: bafy_schedule_cid_001

### Test Sessions

| Session ID | State | Purpose |
|------------|-------|---------|
| sess_123 | Valid | Standard test session |
| sess_456 | Valid | User with all scopes |
| sess_expired | Expired | Test expiration |
| sess_revoked | Revoked | Test revocation |
| sess_exhausted | Zero uses | Test usage limit |

## Provider Interfaces

### Session Provider

```go
type SessionProvider interface {
    ValidateSession(ctx context.Context, req *models.ValidateSessionRequest) error
}
```

### Category Index Provider

```go
type CategoryIndexProvider interface {
    GetCategoryIndex(ctx context.Context, participantID string) (*models.CategoryIndex, error)
}
```

### Filesystem Provider

```go
type FilesystemProvider interface {
    FetchByCID(ctx context.Context, cid string) (*models.FileContent, error)
}
```

### LLM Provider

```go
type LLMProvider interface {
    Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
    GetModelName() string
}
```

## Running Tests

```bash
go test ./...
```

With coverage:
```bash
go test -cover ./...
```

Verbose output:
```bash
go test -v ./...
```

## Project Structure

```
access-bot/
├── cmd/
│   └── main.go                     # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration loading
│   │   └── config_test.go
│   ├── domain/
│   │   └── categories.go           # Domain types and constants
│   ├── handlers/
│   │   ├── answer.go               # HTTP handlers
│   │   ├── answer_test.go
│   │   └── router.go               # HTTP router setup
│   ├── logging/
│   │   └── logger.go               # Structured logging
│   ├── mocks/
│   │   ├── session_provider.go     # Mock session validation
│   │   ├── categoryindex_provider.go
│   │   ├── filesystem_provider.go
│   │   ├── llm_provider.go
│   │   └── providers_test.go
│   ├── models/
│   │   ├── request.go              # Request structs
│   │   ├── response.go             # Response structs
│   │   └── session.go              # Domain models
│   ├── prompts/
│   │   ├── builder.go              # Prompt construction
│   │   └── builder_test.go
│   ├── providers/
│   │   ├── session.go              # SessionProvider interface
│   │   ├── categoryindex.go
│   │   ├── filesystem.go
│   │   └── llm.go
│   ├── services/
│   │   ├── accesslayer/
│   │   │   ├── service.go          # Core business logic
│   │   │   └── service_test.go
│   │   ├── filesystem/
│   │   │   └── pinata.go           # Pinata IPFS client
│   │   ├── llm/
│   │   │   ├── claude.go           # Claude API client
│   │   │   └── gemini.go           # Gemini API client
│   │   └── session/
│   │       └── intelligence.go     # Intelligence service client
│   └── validation/
│       ├── request.go              # Request validation
│       └── request_test.go
├── go.mod
├── .env.example
└── README.md
```

## Design Decisions

1. **Standard library first**: Uses net/http, log/slog, and encoding/json
2. **Interface-based providers**: All external dependencies are behind interfaces
3. **Delegated session validation**: Session validation is handled by participant-intelligence-service
4. **No markdown transformation**: v1 passes full markdown to the LLM
5. **Safe failure by default**: Any error results in a generic safe message
6. **Reason codes for debugging**: Machine-readable reasons in response for diagnostics
7. **No secrets in logs**: Raw markdown content and API keys are not logged

## Future Enhancements (Not in v1)

- [ ] Markdown parsing into structured fields
- [ ] Derived reduced views based on permissions
- [ ] Downstream agent routing
- [ ] Action execution
- [ ] Database-backed category index storage
- [ ] Rate limiting
- [ ] Metrics and tracing