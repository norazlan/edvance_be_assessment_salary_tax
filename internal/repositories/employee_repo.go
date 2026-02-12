package repositories

import (
	"database/sql"
	"fmt"

	"edvance-assessment/internal/models"
)

// EmployeeRepository handles database operations for employees
type EmployeeRepository struct {
	DB *sql.DB
}

// NewEmployeeRepository creates a new repository instance
func NewEmployeeRepository(db *sql.DB) *EmployeeRepository {
	return &EmployeeRepository{DB: db}
}

// Upsert inserts a new employee or updates if name already exists
func (r *EmployeeRepository) Upsert(name string, annualSalary, monthlyIncomeTax float64) error {
	_, err := r.DB.Exec(`
		INSERT INTO employee (name, annual_salary, monthly_income_tax)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET annual_salary = $2,
		    monthly_income_tax = $3,
		    updated_at = NOW()
	`, name, annualSalary, monthlyIncomeTax)
	if err != nil {
		return fmt.Errorf("failed to upsert employee %s: %w", name, err)
	}
	return nil
}

// GetAll fetches all employees from the database
func (r *EmployeeRepository) GetAll() ([]models.EmployeeResponse, error) {
	rows, err := r.DB.Query(`
		SELECT
			TO_CHAR(updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS time_stamp,
			name,
			annual_salary,
			monthly_income_tax
		FROM employee
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees: %w", err)
	}
	defer rows.Close()

	var employees []models.EmployeeResponse
	for rows.Next() {
		var e models.EmployeeResponse
		if err := rows.Scan(&e.TimeStamp, &e.EmployeeName, &e.AnnualSalary, &e.MonthlyIncomeTax); err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return employees, nil
}
