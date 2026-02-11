package repositories

import (
	"database/sql"
	"fmt"
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
