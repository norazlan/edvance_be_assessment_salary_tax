package domains

// TaxCalculator is the contract for the tax calculation strategy
type TaxCalculator interface {
	CalculateAnnualTax(salary float64) float64
}

// ProgressiveTaxStrategy is a concrete implementation for progressive tax calculation
type ProgressiveTaxStrategy struct {
	Brackets []TaxBracket
}

type TaxBracket struct {
	Min, Max, Rate float64
}

func (s *ProgressiveTaxStrategy) CalculateAnnualTax(salary float64) float64 {
	totalTax := 0.0
	for _, b := range s.Brackets {
		if salary > b.Min {
			upper := b.Max
			if salary < b.Max {
				upper = salary
			}
			totalTax += (upper - b.Min) * b.Rate
		}
	}
	return totalTax
}
