package handlers

import (
	"edvance-assessment/internal/models"
	"edvance-assessment/internal/repositories"
	"edvance-assessment/internal/services"
	"edvance-assessment/pkg"

	"github.com/gofiber/fiber/v3"
)

type PayslipHandler struct {
	Service    *services.PayslipService
	EmployeeRepo *repositories.EmployeeRepository
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

	payslipData := h.Service.Generate_monthly_payslip(req.Name, req.Salary)

	// Save to database (upsert)
	if err := h.EmployeeRepo.Upsert(req.Name, req.Salary, payslipData.MonthlyIncomeTax); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to save employee data",
			Errors:  []string{err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Success: true,
		Message: "Monthly payslip generated successfully",
		Data:    payslipData,
	})
}
