-- Employee table to store payslip records
CREATE TABLE IF NOT EXISTS employee (
    id                SERIAL        PRIMARY KEY,
    name              VARCHAR(100)  NOT NULL UNIQUE,
    annual_salary     NUMERIC(15,2) NOT NULL,
    monthly_income_tax NUMERIC(15,2) NOT NULL,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
