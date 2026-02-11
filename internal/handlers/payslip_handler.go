package handlers

import (
	"edvance-assessment/internal/models"
	"edvance-assessment/pkg"

	"github.com/gofiber/fiber/v3"
)

type PayslipHandler struct{}

func NewPayslipHandler() *PayslipHandler {
	return &PayslipHandler{}
}

func (h *PayslipHandler) GenMonthlyPayslip(c fiber.Ctx) error {
	var req models.PayslipRequest

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Errors:  []string{err.Error()},
		})
	}

	validationErrors := pkg.ValidateStruct(req)
	if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  validationErrors,
		})
	}

	payslipData := map[string]interface{}{
		"name":           req.Name,
		"salary":         req.Salary,
		"monthly_salary": req.Salary,
	}

	return c.Status(fiber.StatusOK).JSON(models.PayslipResponse{
		Success: true,
		Message: "Monthly payslip generated successfully",
		Data:    payslipData,
	})
}
