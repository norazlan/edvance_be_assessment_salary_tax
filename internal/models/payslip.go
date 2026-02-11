package models

type PayslipRequest struct {
	Name   string  `json:"name" validate:"required,min=2,max=100"`
	Salary float64 `json:"salary" validate:"required,gt=0"`
}

type PayslipResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}
