-- Tax Brackets table with versioning support
CREATE TABLE IF NOT EXISTS tax_brackets (
    id             SERIAL        PRIMARY KEY,
    version        INT           NOT NULL DEFAULT 1,
    effective_date DATE          NOT NULL DEFAULT CURRENT_DATE,
    min_salary     NUMERIC(15,2) NOT NULL,
    max_salary     NUMERIC(15,2) NOT NULL,
    rate           NUMERIC(5,4)  NOT NULL,
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_min_lt_max  CHECK (min_salary < max_salary),
    CONSTRAINT chk_rate_range  CHECK (rate >= 0 AND rate <= 1)
);

-- Index for fast lookup of active brackets by version
CREATE INDEX IF NOT EXISTS idx_tax_brackets_active_version
    ON tax_brackets (version, is_active)
    WHERE is_active = TRUE;

-- Seed initial brackets (version 1)
INSERT INTO tax_brackets (version, effective_date, min_salary, max_salary, rate)
VALUES
    (1, '2026-01-01', 0,      20000,     0.0000),
    (1, '2026-01-01', 20001,  40000,     0.1000),
    (1, '2026-01-01', 40001,  80000,     0.2000),
    (1, '2026-01-01', 80001,  180000,    0.3000),
    (1, '2026-01-01', 180001, 999999999, 0.4000);
