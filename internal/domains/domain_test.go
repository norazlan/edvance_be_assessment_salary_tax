package domains

import "testing"

func TestProgressiveTaxStrategy_CalculateAnnualTax(t *testing.T) {
	strategy := &ProgressiveTaxStrategy{
		Brackets: []TaxBracket{
			{Min: 0, Max: 20000, Rate: 0.0},
			{Min: 20001, Max: 40000, Rate: 0.1},
			{Min: 40001, Max: 80000, Rate: 0.2},
			{Min: 80001, Max: 180000, Rate: 0.3},
			{Min: 180001, Max: 999999999, Rate: 0.4},
		},
	}

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
