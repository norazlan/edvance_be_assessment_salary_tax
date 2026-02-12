package handlers

import (
	"edvance-assessment/internal/models"
	"edvance-assessment/internal/repositories"
	"edvance-assessment/internal/services"
	"edvance-assessment/pkg"

	"github.com/gofiber/fiber/v3"
)

type PayslipHandler struct {
	Service      *services.PayslipService
	EmployeeRepo *repositories.EmployeeRepository
	EmailService *services.EmailService
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

func (h *PayslipHandler) GetAllEmployees(c fiber.Ctx) error {
	employees, err := h.EmployeeRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to fetch employees",
			Errors:  []string{err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.EmployeeListResponse{
		Success: true,
		Message: "All Monthly payslip fetched successfully",
		Data:    employees,
	})
}

func (h *PayslipHandler) SendEmail(c fiber.Ctx) error {
	var req models.SendEmailRequest

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

	// Fetch all employees
	employees, err := h.EmployeeRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to fetch employees",
			Errors:  []string{err.Error()},
		})
	}

	if len(employees) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "No employee data available to send",
			Errors:  []string{"employee list is empty"},
		})
	}

	// Generate CSV
	csvData, err := services.GenerateCSV(employees)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to generate CSV",
			Errors:  []string{err.Error()},
		})
	}

	// Send email
	if err := h.EmailService.SendEmployeeReport(req.Email, csvData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to send email",
			Errors:  []string{err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Success: true,
		Message: "Employee payslip report sent successfully to " + req.Email,
		Data:    nil,
	})
}
