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

// DeactivateAllVersions marks all active brackets as inactive
func (r *TaxBracketRepository) DeactivateAllVersions() error {
	_, err := r.DB.Exec(`
UPDATE tax_brackets
SET is_active = FALSE, updated_at = NOW()
WHERE is_active = TRUE
`)
	if err != nil {
		return fmt.Errorf("failed to deactivate all versions: %w", err)
	}
	return nil
}

// ActivateVersion activates a specific version (deactivates all others first)
func (r *TaxBracketRepository) ActivateVersion(version int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate all active versions
	_, err = tx.Exec(`
UPDATE tax_brackets
SET is_active = FALSE, updated_at = NOW()
WHERE is_active = TRUE
`)
	if err != nil {
		return fmt.Errorf("failed to deactivate all versions: %w", err)
	}

	// Activate the specified version
	result, err := tx.Exec(`
UPDATE tax_brackets
SET is_active = TRUE, updated_at = NOW()
WHERE version = $1
`, version)
	if err != nil {
		return fmt.Errorf("failed to activate version %d: %w", version, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("version %d not found", version)
	}

	return tx.Commit()
}

// VersionExists checks if a version exists in the database
func (r *TaxBracketRepository) VersionExists(version int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(`
SELECT EXISTS(SELECT 1 FROM tax_brackets WHERE version = $1)
`, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check version existence: %w", err)
	}
	return exists, nil
}

// GetMaxVersion returns the highest version number across all brackets (active or not)
func (r *TaxBracketRepository) GetMaxVersion() (int, error) {
	var version int
	err := r.DB.QueryRow(`
SELECT COALESCE(MAX(version), 0)
FROM tax_brackets
`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get max version: %w", err)
	}
	return version, nil
}

// InsertBrackets inserts a new set of tax brackets with a given version
func (r *TaxBracketRepository) InsertBrackets(version int, effectiveDate string, isActive bool, brackets []struct{ Min, Max, Rate float64 }) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, b := range brackets {
		_, err := tx.Exec(`
INSERT INTO tax_brackets (version, effective_date, min_salary, max_salary, rate, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
`, version, effectiveDate, b.Min, b.Max, b.Rate, isActive)
		if err != nil {
			return fmt.Errorf("failed to insert tax bracket: %w", err)
		}
	}

	return tx.Commit()
}
