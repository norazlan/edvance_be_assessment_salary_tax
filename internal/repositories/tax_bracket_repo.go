package repositories

import (
	"database/sql"
	"fmt"

	"edvance-assessment/internal/domains"
)

// TaxBracketRepository handles database operations for tax brackets
type TaxBracketRepository struct {
	DB *sql.DB
}

// NewTaxBracketRepository creates a new repository instance
func NewTaxBracketRepository(db *sql.DB) *TaxBracketRepository {
	return &TaxBracketRepository{DB: db}
}

// GetActiveByVersion fetches active tax brackets for a specific version
func (r *TaxBracketRepository) GetActiveByVersion(version int) ([]domains.TaxBracket, error) {
	rows, err := r.DB.Query(`
SELECT min_salary, max_salary, rate
FROM tax_brackets
WHERE version = $1 AND is_active = TRUE
ORDER BY min_salary ASC
`, version)
	if err != nil {
		return nil, fmt.Errorf("failed to query tax brackets: %w", err)
	}
	defer rows.Close()

	var brackets []domains.TaxBracket
	for rows.Next() {
		var b domains.TaxBracket
		if err := rows.Scan(&b.Min, &b.Max, &b.Rate); err != nil {
			return nil, fmt.Errorf("failed to scan tax bracket: %w", err)
		}
		brackets = append(brackets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return brackets, nil
}

// GetLatestVersion returns the latest active version number
func (r *TaxBracketRepository) GetLatestVersion() (int, error) {
	var version int
	err := r.DB.QueryRow(`
SELECT COALESCE(MAX(version), 0)
FROM tax_brackets
WHERE is_active = TRUE
`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}
	return version, nil
}

// GetActiveBrackets fetches the latest active tax brackets
func (r *TaxBracketRepository) GetActiveBrackets() ([]domains.TaxBracket, int, error) {
	version, err := r.GetLatestVersion()
	if err != nil {
		return nil, 0, err
	}

	brackets, err := r.GetActiveByVersion(version)
	if err != nil {
		return nil, 0, err
	}

	return brackets, version, nil
}

// DeactivateVersion marks all brackets of a version as inactive
func (r *TaxBracketRepository) DeactivateVersion(version int) error {
	_, err := r.DB.Exec(`
UPDATE tax_brackets
SET is_active = FALSE, updated_at = NOW()
WHERE version = $1
`, version)
	if err != nil {
		return fmt.Errorf("failed to deactivate version %d: %w", version, err)
	}
	return nil
}
