# Access Layer Service (v1)

A participant-centric access layer service that validates session permissions, fetches authorized participant data, and uses Claude to answer questions within the approved context.

## Overview

The Access Layer Service acts as a policy boundary between requesting agents (like OpenClaw WhatsApp bot) and participant data. It:

1. Validates session access for requested data categories
2. Fetches markdown files by CID for approved categories
3. Sends the user question plus approved context to Claude
4. Returns Claude's answer or a safe failure response

**Key principle**: The access layer enforces all permission checks. The LLM is only a reasoning engine—it is never trusted for permission decisions.

## Features

- **Session Validation**: Verifies session existence, ownership, expiration, and revocation
- **Scope-Based Permissions**: Maps data categories to OAuth-style scopes
- **Content Retrieval**: Fetches markdown content from a CID-based filesystem
- **LLM Integration**: Sends structured prompts to Claude API
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
  "status": "cannot_answer",
  "answer": "Sorry, I can't answer that.",
  "used_categories": [],
  "llm_model": null,
  "reason": "insufficient_authorized_context"
}
```

### Status Values
- `ok` - Request processed successfully
- `cannot_answer` - Unable to answer safely (session issues, missing data, LLM failure)
- `permission_denied` - Session lacks required permissions
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
| PAYMENT | finance.payment.read |

A session must have all required scopes to access the requested categories.

## Session Validation Rules

Before fetching data or calling the LLM, the service validates:

1. Session exists
2. Session belongs to the requested participant_id
3. Session is bound to the requesting agent_id
4. Session is not expired
5. Session is not revoked
6. Usage limit is not exceeded
7. All required scopes are authorized

## Running Locally

### Prerequisites
- Go 1.22+
- Anthropic API key (optional for mock mode)

### Setup

1. Clone the repository:
```bash
cd /path/to/access-bot
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Configure `.env` (optional):
```bash
# For real Claude API:
ANTHROPIC_API_KEY=sk-ant-api03-xxxxx

# Or leave empty to use mock LLM
```

4. Run the server:
```bash
go run cmd/server/main.go
```

5. Server starts at `http://localhost:8080`

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | HTTP server port |
| HOST | localhost | HTTP server host |
| ANTHROPIC_API_KEY | (empty) | Claude API key (uses mock if empty) |
| CLAUDE_MODEL | claude-sonnet-4-20250514 | Claude model to use |
| HTTP_CLIENT_TIMEOUT | 30 | HTTP client timeout in seconds |
| LOG_LEVEL | info | Log level (debug, info, warn, error) |

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

The v1 implementation uses in-memory mocks for session, category index, and filesystem providers.

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
- PAYMENT: bafy_payment_cid_001

### Test Sessions

| Session ID | State | Purpose |
|------------|-------|---------|
| sess_123 | Valid | Standard test session |
| sess_456 | Valid | User with all scopes |
| sess_expired | Expired | Test expiration |
| sess_revoked | Revoked | Test revocation |
| sess_exhausted | Zero uses | Test usage limit |

## Plugging in Real Providers

### Session Provider

Implement the `SessionProvider` interface:

```go
type SessionProvider interface {
    GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}
```

### Category Index Provider

Implement the `CategoryIndexProvider` interface:

```go
type CategoryIndexProvider interface {
    GetCategoryIndex(ctx context.Context, participantID string) (*models.CategoryIndex, error)
}
```

### Filesystem Provider

Implement the `FilesystemProvider` interface:

```go
type FilesystemProvider interface {
    FetchByCID(ctx context.Context, cid string) (*models.FileContent, error)
}
```

### LLM Provider

Implement the `LLMProvider` interface or use the included Claude client:

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
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration loading
│   │   └── config_test.go
│   ├── domain/
│   │   └── categories.go        # Domain types and constants
│   ├── handlers/
│   │   ├── answer.go            # HTTP handlers
│   │   ├── answer_test.go
│   │   └── router.go            # HTTP router setup
│   ├── logging/
│   │   └── logger.go            # Structured logging
│   ├── mocks/
│   │   ├── session_provider.go
│   │   ├── categoryindex_provider.go
│   │   ├── filesystem_provider.go
│   │   ├── llm_provider.go
│   │   └── providers_test.go
│   ├── models/
│   │   ├── request.go           # Request structs
│   │   ├── response.go          # Response structs
│   │   └── session.go           # Domain models
│   ├── prompts/
│   │   ├── builder.go           # Prompt construction
│   │   └── builder_test.go
│   ├── providers/
│   │   ├── session.go           # Provider interfaces
│   │   ├── categoryindex.go
│   │   ├── filesystem.go
│   │   └── llm.go
│   ├── services/
│   │   ├── accesslayer/
│   │   │   ├── service.go       # Core business logic
│   │   │   └── service_test.go
│   │   └── llm/
│   │       └── claude.go        # Claude API client
│   └── validation/
│       ├── request.go           # Request validation
│       └── request_test.go
├── go.mod
├── .env.example
└── README.md
```

## Design Decisions

1. **Standard library first**: Uses net/http, log/slog, and encoding/json
2. **Interface-based providers**: All external dependencies are behind interfaces
3. **No markdown transformation**: v1 passes full markdown to Claude
4. **Safe failure by default**: Any error results in a generic safe message
5. **Reason codes for debugging**: Machine-readable reasons in response for diagnostics
6. **No secrets in logs**: Raw markdown content is not logged

## Future Enhancements (Not in v1)

- [ ] Markdown parsing into structured fields
- [ ] Derived reduced views based on permissions
- [ ] Downstream agent routing
- [ ] Action execution
- [ ] Database-backed session storage
- [ ] Rate limiting
- [ ] Metrics and tracing
