# Copilot Instructions for Edvance BE Assessment

## Architecture

Go REST API using **Fiber v3** with clean architecture. Data flows:

```
PostgreSQL → Repository → GOB cache (data/tax_brackets.gob) → Strategy → Service → Handler
```

```
app.go                  # Entry point - wiring, middleware, routes, graceful shutdown
config/                 # Env config, DB connection, migration runner
migrations/             # Ordered SQL migration files (001_*.sql, ...)
internal/
  domains/              # Domain entities + strategy interfaces (TaxCalculator)
  repositories/         # Database access layer (TaxBracketRepository, EmployeeRepository)
  services/             # Business logic (PayslipService, EmailService)
  handlers/             # HTTP handler (PayslipHandler) + CLI handler (CLIHandler)
  models/               # Request/response DTOs
pkg/                    # Shared utilities (validator, GOB storage, file watcher)
data/                   # Runtime GOB cache (ephemeral, deleted on shutdown)
```

## Key Patterns

### Strategy Pattern (Domain Layer)

`TaxCalculator` interface in `internal/domains/` allows swappable tax algorithms:

```go
type TaxCalculator interface {
    CalculateAnnualTax(salary float64) float64
}
```

`ProgressiveTaxStrategy` is the concrete implementation with inclusive bracket ranges (`upper - b.Min + 1`).

### Dependency Injection (Constructor-Based)

All wiring happens in `app.go main()`:

```go
repo := repositories.NewTaxBracketRepository(db)
employeeRepo := repositories.NewEmployeeRepository(db)
strategy := &domains.ProgressiveTaxStrategy{Brackets: brackets}
service := services.NewPayslipService(strategy)
emailService := services.NewEmailService(services.SMTPConfig{...})
handler := &handlers.PayslipHandler{Service: service, EmployeeRepo: employeeRepo, EmailService: emailService}
```

### Repository Pattern

`TaxBracketRepository` wraps `*sql.DB`. Key methods: `GetActiveBrackets()`, `GetActiveByVersion(v)`, `GetLatestVersion()`, `DeactivateVersion(v)`.

`EmployeeRepository` wraps `*sql.DB`. Key method: `Upsert(name, annualSalary, monthlyIncomeTax)` — uses `INSERT ... ON CONFLICT (name) DO UPDATE` for upsert.

### Dual Mode (Web / CLI)

App runs as web server or CLI based on `APP_MODE` env variable:

- `web` (default) — Fiber HTTP server with `POST /gen_monthly_payslip`, `GET /employees`, `POST /send_email`
- `cli` — Prompts user for `name` and `annual salary`, prints payslip to console

Both modes use the same `PayslipService`, validation via `pkg.ValidateStruct()`, and save employee data to DB via `EmployeeRepository.Upsert()`.

### Employee Table (Database)

`employee` table stores payslip records with `name` as UNIQUE constraint:

- New name → INSERT
- Existing name → UPDATE `annual_salary`, `monthly_income_tax`, `updated_at`

### Tax Bracket Versioning (Database)

`tax_brackets` table supports versioned, soft-deletable brackets:

- `version` (int) — group brackets by version number
- `effective_date` — when the version takes effect
- `is_active` (bool) — soft delete; `DeactivateVersion()` sets FALSE
- Partial index on `(version, is_active) WHERE is_active = TRUE`

### GOB File Cache + Hot Reload

On startup: DB brackets → `pkg.SaveToGob()` → `data/tax_brackets.gob`. A `fsnotify` watcher goroutine auto-reloads brackets when the GOB file changes. File is deleted on graceful shutdown.

### Email Service (AWS SES SMTP)

`EmailService` in `internal/services/email.go` sends emails via AWS SES SMTP using Go's `net/smtp` package:

- `SMTPConfig` — holds `APIKey`, `Host`, `User`, `Password`, `MailFrom`
- `NewEmailService(cfg SMTPConfig)` — creates service with SMTP credentials
- `GenerateCSV(employees)` — converts `[]EmployeeResponse` to CSV bytes (Timestamp, Employee Name, Annual Salary, Monthly Income Tax)
- `SendEmployeeReport(toEmail, csvData)` — builds MIME multipart message with CSV attachment, sends via `smtp.SendMail()` with `smtp.PlainAuth`

`POST /send_email` endpoint: validates `SendEmailRequest` (email field), fetches all employees, generates CSV, sends via SMTP.

### Request/Response Models

In `internal/models/`:

- `PayslipRequest` — `json` + `validate` tags
- `PayslipResponse` — `EmployeeName`, `GrossMonthlyIncome`, `MonthlyIncomeTax`, `NetMonthlyIncome`
- `SuccessResponse` — envelope: `Success`, `Message`, `Data` (`interface{}`)
- `ErrorResponse` — envelope: `Success`, `Message`, `Errors[]`
- `EmployeeResponse` — `TimeStamp`, `EmployeeName`, `AnnualSalary`, `MonthlyIncomeTax`
- `EmployeeListResponse` — envelope: `Success`, `Message`, `Data` (`[]EmployeeResponse`)
- `SendEmailRequest` — `Email` with `required,email` validation

### Validation

Use `pkg.ValidateStruct()` (go-playground/validator v10). Human-readable error messages for `required`, `min`, `max`, `gt` tags.

### Monetary Precision

All money values rounded to 2 decimal places: `math.Round(x*100) / 100`.

## Development

- **Run**: `go run app.go` (requires `.env` — copy from `.env.sample`, and PostgreSQL running)
- **Test**: `go test -v ./...`
- **Lint**: `golangci-lint run`
- **Migrations**: Auto-run on startup via `config.RunMigrations()` — add new `.sql` files in `migrations/`

## Configuration (`.env`)

| Variable               | Default       | Purpose                               |
| ---------------------- | ------------- | ------------------------------------- |
| `APP_ENV`              | `development` | `production` enables prefork          |
| `APP_PORT`             | `3000`        | Server port                           |
| `APP_MODE`             | `web`         | `web` or `cli`                        |
| `DB_HOST`              | `localhost`   | PostgreSQL host                       |
| `DB_PORT`              | `5432`        | PostgreSQL port                       |
| `DB_USER`              | `postgres`    | PostgreSQL user                       |
| `DB_PASS`              | _(empty)_     | PostgreSQL password                   |
| `DB_NAME`              | `edvance`     | PostgreSQL database                   |
| `EMAIL_SERVER_API_KEY` | _(empty)_     | Email server API key                  |
| `EMAIL_SMTP_HOST`      | _(empty)_     | SMTP host with port (e.g. `host:587`) |
| `EMAIL_SMTP_USER`      | _(empty)_     | SMTP username (AWS SES SMTP user)     |
| `EMAIL_SMTP_PASS`      | _(empty)_     | SMTP password (AWS SES SMTP password) |
| `EMAIL_SMTP_MAIL_FROM` | _(empty)_     | Verified sender email address         |

## Adding New Features

1. Define request/response models in `internal/models/`
2. Add domain entities/interfaces in `internal/domains/`
3. Add repository in `internal/repositories/` if DB access needed
4. Add service in `internal/services/` with domain interface dependency
5. Create handler in `internal/handlers/`
6. Wire dependencies and add route in `app.go`
7. Add SQL migration in `migrations/` (numbered, e.g. `002_*.sql`)
