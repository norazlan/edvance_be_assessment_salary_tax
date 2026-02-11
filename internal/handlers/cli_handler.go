package handlers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"edvance-assessment/internal/models"
	"edvance-assessment/internal/repositories"
	"edvance-assessment/internal/services"
	"edvance-assessment/pkg"
)

// CLIHandler handles the CLI mode for payslip generation
type CLIHandler struct {
	Service      *services.PayslipService
	EmployeeRepo *repositories.EmployeeRepository
}

// Run prompts user for input and displays the payslip
func (h *CLIHandler) Run() {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for name
	fmt.Print("Enter employee name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Prompt for annual salary
	fmt.Print("Enter annual salary: ")
	salaryStr, _ := reader.ReadString('\n')
	salaryStr = strings.TrimSpace(salaryStr)

	salary, err := strconv.ParseFloat(salaryStr, 64)
	if err != nil {
		fmt.Println("\nError: Salary must be a valid number")
		return
	}

	// Validate using the same PayslipRequest model
	req := models.PayslipRequest{
		Name:   name,
		Salary: salary,
	}

	validationErrors := pkg.ValidateStruct(req)
	if len(validationErrors) > 0 {
		fmt.Println("\nValidation errors:")
		for _, e := range validationErrors {
			fmt.Printf("  - %s\n", e)
		}
		return
	}

	// Generate payslip
	payslip := h.Service.Generate_monthly_payslip(name, salary)

	// Save to database (upsert)
	if err := h.EmployeeRepo.Upsert(name, salary, payslip.MonthlyIncomeTax); err != nil {
		fmt.Printf("\nWarning: Failed to save employee data: %v\n", err)
	}

	// Print formatted output
	fmt.Println()
	fmt.Printf(" Monthly Payslip for: \"%s\"\n", payslip.EmployeeName)
	fmt.Printf(" Gross Monthly Income: $%.2f\n", payslip.GrossMonthlyIncome)
	fmt.Printf(" Monthly Income Tax: $%.2f\n", payslip.MonthlyIncomeTax)
	fmt.Printf(" Net Monthly Income: $%.2f\n", payslip.NetMonthlyIncome)
}
