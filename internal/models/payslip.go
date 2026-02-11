package models

type PayslipRequest struct {
	Name   string  `json:"name" validate:"required,min=2,max=100"`
	Salary float64 `json:"salary" validate:"required,gt=0"`
}

type PayslipResponse struct {
	EmployeeName       string  `json:"employee_name"`
	GrossMonthlyIncome float64 `json:"gross_monthly_income"`
	MonthlyIncomeTax   float64 `json:"monthly_income_tax"`
	NetMonthlyIncome   float64 `json:"net_monthly_income"`
}

type ErrorResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

type SuccessResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    PayslipResponse `json:"data"`
}
