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
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type EmployeeResponse struct {
	TimeStamp        string  `json:"time_stamp"`
	EmployeeName     string  `json:"employee_name"`
	AnnualSalary     float64 `json:"annual_salary"`
	MonthlyIncomeTax float64 `json:"monthly_income_tax"`
}

type EmployeeListResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    []EmployeeResponse `json:"data"`
}

type SendEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type TaxBracketInput struct {
	Min  float64 `json:"min" validate:"gte=0"`
	Max  float64 `json:"max" validate:"required,gtfield=Min"`
	Rate float64 `json:"rate" validate:"gte=0,lte=1"`
}

type SetTaxBracketsRequest struct {
	Brackets      []TaxBracketInput `json:"brackets" validate:"required,min=0,dive"`
	EffectiveDate string            `json:"effective_date" validate:"omitempty"`
	IsActive      *bool             `json:"is_active"`
}

type SetTaxBracketsResponse struct {
	Version       int               `json:"version"`
	EffectiveDate string            `json:"effective_date"`
	IsActive      bool              `json:"is_active"`
	Brackets      []TaxBracketInput `json:"brackets"`
}
