package services

import (
	"edvance-assessment/internal/domains"
	"edvance-assessment/internal/models"
	"math"
)

type PayslipService struct {
	Calculator domains.TaxCalculator
}

func NewPayslipService(calc domains.TaxCalculator) *PayslipService {
	return &PayslipService{Calculator: calc}
}

func (s *PayslipService) Generate_monthly_payslip(name string, annualSalary float64) models.PayslipResponse {
	annualTax := s.Calculator.CalculateAnnualTax(annualSalary)

	grossMonthly := annualSalary / 12
	monthlyTax := annualTax / 12

	return models.PayslipResponse{
		EmployeeName:       name,
		GrossMonthlyIncome: math.Round(grossMonthly*100) / 100,
		MonthlyIncomeTax:   math.Round(monthlyTax*100) / 100,
		NetMonthlyIncome:   math.Round((grossMonthly-monthlyTax)*100) / 100,
	}
}
