package handlers

import (
	"fmt"
	"log"

	"edvance-assessment/internal/domains"
	"edvance-assessment/internal/models"
	"edvance-assessment/internal/repositories"
	"edvance-assessment/internal/services"
	"edvance-assessment/pkg"

	"github.com/gofiber/fiber/v3"
)

type PayslipHandler struct {
	Service        *services.PayslipService
	EmployeeRepo   *repositories.EmployeeRepository
	EmailService   *services.EmailService
	TaxBracketRepo *repositories.TaxBracketRepository
	Strategy       *domains.ProgressiveTaxStrategy
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

func (h *PayslipHandler) SetTaxBrackets(c fiber.Ctx) error {
	var req models.SetTaxBracketsRequest

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

	// Determine next version
	maxVersion, err := h.TaxBracketRepo.GetMaxVersion()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to get current version",
			Errors:  []string{err.Error()},
		})
	}
	newVersion := maxVersion + 1

	// Default is_active to false if not provided
	isActive := false
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Default effective_date to today if not provided
	effectiveDate := req.EffectiveDate
	if effectiveDate == "" {
		effectiveDate = "now()"
	}

	// Build bracket params
	bracketParams := make([]struct{ Min, Max, Rate float64 }, len(req.Brackets))
	bracketInputs := make([]models.TaxBracketInput, len(req.Brackets))
	for i, b := range req.Brackets {
		bracketParams[i] = struct{ Min, Max, Rate float64 }{Min: b.Min, Max: b.Max, Rate: b.Rate}
		bracketInputs[i] = b
	}

	// Insert into database
	if err := h.TaxBracketRepo.InsertBrackets(newVersion, effectiveDate, isActive, bracketParams); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to insert tax brackets",
			Errors:  []string{err.Error()},
		})
	}

	// If new brackets are active, deactivate old versions and refresh GOB cache
	if isActive {
		// Deactivate all other versions
		if err := h.TaxBracketRepo.DeactivateAllVersions(); err != nil {
			log.Printf("Warning: Failed to deactivate old versions: %v\n", err)
		}

		// Re-activate only the new version (since DeactivateAllVersions deactivated it too)
		if err := h.TaxBracketRepo.ActivateVersion(newVersion); err != nil {
			log.Printf("Warning: Failed to re-activate new version: %v\n", err)
		}

		// Refresh GOB cache
		activeBrackets, _, err := h.TaxBracketRepo.GetActiveBrackets()
		if err != nil {
			log.Printf("Warning: Failed to reload active brackets after insert: %v\n", err)
		} else {
			// Update strategy directly (don't rely on file watcher)
			h.Strategy.Brackets = activeBrackets
			log.Printf("Strategy updated with new active brackets (version %d)\n", newVersion)

			if err := pkg.SaveToGob("data/tax_brackets.gob", activeBrackets); err != nil {
				log.Printf("Warning: Failed to update GOB cache: %v\n", err)
			} else {
				log.Printf("GOB cache updated with new active brackets (version %d)\n", newVersion)
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Tax brackets version %d created successfully", newVersion),
		Data: models.SetTaxBracketsResponse{
			Version:       newVersion,
			EffectiveDate: effectiveDate,
			IsActive:      isActive,
			Brackets:      bracketInputs,
		},
	})
}

func (h *PayslipHandler) SetTaxBracketsActive(c fiber.Ctx) error {
	var req models.SetTaxBracketsActiveRequest

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

	// Check if version exists
	exists, err := h.TaxBracketRepo.VersionExists(req.Version)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to check version",
			Errors:  []string{err.Error()},
		})
	}
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("Tax brackets version %d not found", req.Version),
			Errors:  []string{"version does not exist"},
		})
	}

	// Activate the version (deactivates all others in a transaction)
	if err := h.TaxBracketRepo.ActivateVersion(req.Version); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to activate version",
			Errors:  []string{err.Error()},
		})
	}

	// Refresh GOB cache
	activeBrackets, _, err := h.TaxBracketRepo.GetActiveBrackets()
	if err != nil {
		log.Printf("Warning: Failed to reload active brackets: %v\n", err)
	} else {
		// Update strategy directly (don't rely on file watcher)
		h.Strategy.Brackets = activeBrackets
		log.Printf("Strategy updated with active brackets (version %d)\n", req.Version)

		if err := pkg.SaveToGob("data/tax_brackets.gob", activeBrackets); err != nil {
			log.Printf("Warning: Failed to update GOB cache: %v\n", err)
		} else {
			log.Printf("GOB cache updated with active brackets (version %d)\n", req.Version)
		}
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Tax brackets version %d activated successfully", req.Version),
		Data:    nil,
	})
}
