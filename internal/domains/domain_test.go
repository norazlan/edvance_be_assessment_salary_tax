package domains_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"edvance-assessment/config"
	"edvance-assessment/internal/domains"
	"edvance-assessment/internal/repositories"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func loadBracketsFromDB(t *testing.T) []domains.TaxBracket {
	t.Helper()

	// Load .env from project root
	_ = godotenv.Load("../../.env")

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "edvance"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	if err := config.RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	repo := repositories.NewTaxBracketRepository(db)
	brackets, _, err := repo.GetActiveBrackets()
	if err != nil {
		t.Fatalf("Failed to load tax brackets: %v", err)
	}

	if len(brackets) == 0 {
		t.Fatal("No active tax brackets found in database")
	}

	return brackets
}

func TestProgressiveTaxStrategy_CalculateAnnualTax(t *testing.T) {
	brackets := loadBracketsFromDB(t)
	strategy := &domains.ProgressiveTaxStrategy{Brackets: brackets}

	tests := []struct {
		name     string
		salary   float64
		expected float64
	}{
		{
			name:     "salary 60000 should return 6000 tax",
			salary:   60000,
			expected: 6000,
		},
		{
			name:     "salary 20000 should return 0 tax",
			salary:   20000,
			expected: 0,
		},
		{
			name:     "salary 20000.5 should return 0 tax",
			salary:   20000.5,
			expected: 0,
		},
		{
			name:     "salary 200000 should return 48000 tax",
			salary:   200000,
			expected: 48000,
		},
		{
			name:     "salary 80150 should return 10045 tax",
			salary:   80150,
			expected: 10045,
		},
		{
			name:     "salary 1000000 should return 368000 tax",
			salary:   1000000,
			expected: 368000,
		},
		{
			name:     "salary 93427.60 should return 14028.28 tax",
			salary:   93427.60,
			expected: 14028.28,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strategy.CalculateAnnualTax(tt.salary)
			if got != tt.expected {
				t.Errorf("CalculateAnnualTax(%v) = %v, want %v", tt.salary, got, tt.expected)
			}
		})
	}
}
