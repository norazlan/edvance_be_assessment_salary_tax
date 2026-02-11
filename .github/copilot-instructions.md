# Copilot Instructions for Edvance BE Assessment

## Architecture Overview

Go REST API using **Fiber v3** framework with clean architecture pattern:

```
app.go              # Entry point - Fiber setup, middleware, routes
config/             # Environment configuration via godotenv
internal/
  handlers/         # HTTP handlers (controller layer)
  models/           # Request/response DTOs
  services/         # Business logic (to be implemented)
  domains/          # Domain entities (to be implemented)
pkg/                # Shared utilities (validation)
```

## Key Patterns

### Handler Pattern

Handlers use constructor pattern and receive Fiber context:

```go
type PayslipHandler struct{}
func NewPayslipHandler() *PayslipHandler { return &PayslipHandler{} }
func (h *PayslipHandler) GenMonthlyPayslip(c fiber.Ctx) error { ... }
```

### Request/Response Models

All models follow consistent structure in `internal/models/`:

- `XxxRequest` - input with `json` + `validate` tags
- `XxxResponse` - success response with `Success`, `Message`, `Data`
- `ErrorResponse` - error response with `Success`, `Message`, `Errors[]`

### Validation

Use `pkg.ValidateStruct()` for request validation (go-playground/validator). Add validation tags directly to request structs:

```go
Name string `json:"name" validate:"required,min=2,max=100"`
```

### Error Handling

Return consistent JSON errors with HTTP status codes:

```go
return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
    Success: false,
    Message: "Validation failed",
    Errors:  validationErrors,
})
```

## Development

- **Run**: `go run app.go` (requires `.env` file - copy from `.env.sample`)
- **Lint**: `golangci-lint run`
- **Format**: Auto-formats on save with `goimports`
- **Test**: `go test -v ./...`

## Configuration

Environment variables via `.env`:

- `APP_ENV` - "development" or "production" (affects prefork)
- `APP_PORT` - Server port (default: 3000)

## Adding New Features

1. Define request/response models in `internal/models/`
2. Create handler in `internal/handlers/` with `NewXxxHandler()` constructor
3. Add route in `app.go`
4. For business logic, add services in `internal/services/`
