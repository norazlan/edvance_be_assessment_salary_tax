# Tax Calculator Payslip Generator

A production-ready Go REST API for generating monthly payslips with progressive tax calculation, employee management, and email reporting.

Built with **Go 1.25**, **Fiber v3**, and **PostgreSQL**.

---

## Features

| Feature                  | Route / Mode                    | Description                                                 |
| ------------------------ | ------------------------------- | ----------------------------------------------------------- |
| Generate Monthly Payslip | `POST /gen_monthly_payslip`     | Calculate gross income, progressive tax, and net income     |
| List All Employees       | `GET /employees`                | Retrieve all employee payslip records                       |
| Send Email Report        | `POST /send_email`              | Email CSV report of all employees via AWS SES SMTP          |
| Set Tax Brackets         | `POST /set_tax_brackets`        | Insert new versioned tax brackets (active or inactive)      |
| Activate Tax Brackets    | `POST /set_tax_brackets_active` | Switch active tax bracket version (hot-swap, zero downtime) |
| CLI Mode                 | `APP_MODE=cli`                  | Interactive terminal-based payslip generation               |
| Auto Migrations          | On startup                      | SQL migrations run automatically                            |
| Graceful Shutdown        | `SIGINT` / `SIGTERM`            | Clean resource teardown, GOB cache cleanup                  |

---

## Architecture

```
PostgreSQL → Repository → GOB Cache → Strategy → Service → Handler
```

```
app.go                    # Entry point — wiring, middleware, routes, graceful shutdown
config/                   # Environment config, DB connection, migration runner
migrations/               # Ordered SQL migrations (001_*.sql, 002_*.sql, ...)
internal/
  domains/                # Domain entities + strategy interfaces
  repositories/           # Database access layer
  services/               # Business logic (PayslipService, EmailService)
  handlers/               # HTTP handlers + CLI handler
  models/                 # Request/response DTOs
pkg/                      # Shared utilities (validator, GOB storage)
data/                     # Runtime GOB cache (ephemeral)
```

### Data Flow

```
Request → Handler (validate) → Service (calculate) → Repository (persist) → Response
                                    ↑
                          Strategy Pattern
                     (ProgressiveTaxStrategy)
```

---

## Design Patterns

### 1. Strategy Pattern

Tax calculation is abstracted behind a `TaxCalculator` interface, allowing swappable algorithms without modifying business logic:

```go
type TaxCalculator interface {
    CalculateAnnualTax(salary float64) float64
}
```

`ProgressiveTaxStrategy` implements inclusive bracket ranges with `math.Round` for 2 decimal place precision. New tax strategies can be plugged in by implementing the interface.

### 2. Repository Pattern

Database access is encapsulated in repository structs (`TaxBracketRepository`, `EmployeeRepository`), providing a clean abstraction over SQL queries. Each repository wraps `*sql.DB` and exposes domain-specific methods.

### 3. Dependency Injection (Constructor-Based)

All dependencies are wired explicitly in `app.go` — no global state, no service locators. This makes the dependency graph visible, testable, and easy to reason about:

```go
strategy := &domains.ProgressiveTaxStrategy{Brackets: brackets}
service  := services.NewPayslipService(strategy)
handler  := &handlers.PayslipHandler{Service: service, ...}
```

### 4. Envelope Response Pattern

All API responses follow a consistent envelope structure:

```json
{ "success": true, "message": "...", "data": { ... } }
{ "success": false, "message": "...", "errors": ["..."] }
```

### 5. Dual Mode (Web / CLI)

The app supports two modes via `APP_MODE` env variable, sharing the same service layer and validation logic — demonstrating separation of concerns between presentation and business logic.

### 6. Versioned Tax Brackets with Hot-Swap

Tax brackets are versioned in the database with soft-delete (`is_active`). Activating a new version deactivates all others in a single transaction, then updates the in-memory strategy directly — **zero downtime, no restart needed**.

---

## Unit Testing Strategy

### Integration Tests with Real Database

Tests in `internal/domains/domain_test.go` use the **actual database** with active tax brackets, ensuring calculations match production data:

```go
func loadBracketsFromDB(t *testing.T) []domains.TaxBracket {
    // Connect to real PostgreSQL, run migrations, load active brackets
}
```

### Table-Driven Tests

Uses Go's idiomatic table-driven test pattern for comprehensive coverage across edge cases:

| Test Case                | Salary    | Expected Tax |
| ------------------------ | --------- | ------------ |
| Standard salary          | 60,000    | 6,000        |
| Below threshold          | 20,000    | 0            |
| Fractional salary (edge) | 20,000.50 | 0            |
| High bracket             | 200,000   | 48,000       |
| Mid-bracket boundary     | 80,150    | 10,045       |
| Very high salary         | 1,000,000 | 368,000      |
| Floating-point precision | 93,427.60 | 14,028.28    |

### Run Tests

```bash
go test -v ./...
```

---

## Build

### Prerequisites

- Go 1.25+
- PostgreSQL (running and accessible)

### Using `build.sh`

The project includes a cross-compilation build script:

```bash
# Build for Linux
./build.sh -l

# Build for Windows
./build.sh -w
```

**Output:** Binary is placed in `./dist/` along with a `.env` file copied from `.env.sample`.

```
dist/
  payslip       # Linux binary
  payslip.exe   # Windows binary
  .env          # Configuration (edit before running)
```

### Manual Build

```bash
go build -o payslip app.go
```

---

## Setup & Run

### 1. Configure Environment

```bash
cp .env.sample .env
# Edit .env with your database and SMTP credentials
```

### 2. Run

```bash
# Web server mode (default)
go run app.go

# CLI mode
APP_MODE=cli go run app.go
```

Migrations run automatically on startup.

### 3. Run with built binary

```bash
./build.sh -l
cd dist
# Edit .env with your credentials
./payslip
```

---

## API Reference

### `POST /gen_monthly_payslip`

Generate a monthly payslip and save employee record.

```json
// Request
{ "name": "John Doe", "salary": 60000 }

// Response (200)
{
  "success": true,
  "message": "Monthly payslip generated successfully",
  "data": {
    "employee_name": "John Doe",
    "gross_monthly_income": 5000.00,
    "monthly_income_tax": 500.00,
    "net_monthly_income": 4500.00
  }
}
```

### `GET /employees`

Retrieve all employee payslip records.

```json
// Response (200)
{
  "success": true,
  "message": "All Monthly payslip fetched successfully",
  "data": [
    {
      "time_stamp": "2026-02-12T10:30:00Z",
      "employee_name": "John Doe",
      "annual_salary": 60000,
      "monthly_income_tax": 500.0
    }
  ]
}
```

### `POST /send_email`

Send CSV report of all employees via email.

```json
// Request
{ "email": "manager@company.com" }

// Response (200)
{
  "success": true,
  "message": "Employee payslip report sent successfully to manager@company.com"
}
```

### `POST /set_tax_brackets`

Insert a new version of tax brackets.

```json
// Request
{
  "brackets": [
    { "min": 0, "max": 25000, "rate": 0 },
    { "min": 25001, "max": 50000, "rate": 0.05 },
    { "min": 50001, "max": 100000, "rate": 0.15 },
    { "min": 100001, "max": 200000, "rate": 0.25 },
    { "min": 200001, "max": 999999999, "rate": 0.35 }
  ],
  "effective_date": "2027-01-01",
  "is_active": false
}
```

- `is_active` defaults to `false` if omitted
- If `is_active: true`, all previous versions are auto-deactivated

### `POST /set_tax_brackets_active`

Activate a specific tax bracket version (deactivates all others).

```json
// Request
{ "version": 2 }

// Response (200)
{
  "success": true,
  "message": "Tax brackets version 2 activated successfully"
}
```

---

## Configuration

| Variable               | Default       | Purpose                               |
| ---------------------- | ------------- | ------------------------------------- |
| `APP_ENV`              | `development` | `production` enables prefork          |
| `APP_PORT`             | `3000`        | Server port                           |
| `APP_MODE`             | `web`         | `web` or `cli`                        |
| `DB_HOST`              | `localhost`   | PostgreSQL host                       |
| `DB_PORT`              | `5432`        | PostgreSQL port                       |
| `DB_USER`              | `postgres`    | PostgreSQL user                       |
| `DB_PASS`              | —             | PostgreSQL password                   |
| `DB_NAME`              | `postgres`    | PostgreSQL database                   |
| `EMAIL_SERVER_API_KEY` | —             | Email server API key                  |
| `EMAIL_SMTP_HOST`      | —             | SMTP host with port (e.g. `host:587`) |
| `EMAIL_SMTP_USER`      | —             | SMTP username                         |
| `EMAIL_SMTP_PASS`      | —             | SMTP password                         |
| `EMAIL_SMTP_MAIL_FROM` | —             | Verified sender email address         |

---

## Technical Highlights

- **Zero-downtime tax bracket switching** — hot-swap active brackets via API without restarting the server
- **Monetary precision** — all financial calculations use `math.Round(x*100) / 100` to avoid floating-point drift
- **Upsert pattern** — employee records use `INSERT ... ON CONFLICT DO UPDATE` for idempotent writes
- **Transactional version switching** — `ActivateVersion()` deactivates all + activates target in a single DB transaction
- **Auto-migrations** — SQL migrations run on startup with a `schema_migrations` tracking table to prevent re-runs
- **Cross-platform builds** — single `build.sh` script for Linux and Windows binaries
- **Graceful shutdown** — signal handling cleans up DB connections and ephemeral cache files
- **MIME multipart email** — CSV attachment built and sent via `net/smtp` with `smtp.PlainAuth`
- **Struct validation** — request validation via `go-playground/validator` with human-readable error messages
- **Clean architecture** — no circular dependencies; clear separation of domains, services, repositories, and handlers
