package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"edvance-assessment/config"
	"edvance-assessment/internal/domains"
	"edvance-assessment/internal/handlers"
	"edvance-assessment/internal/repositories"
	"edvance-assessment/internal/services"
	"edvance-assessment/pkg"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

const TaxBracketsFile = "data/tax_brackets.gob"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	// Connect to database
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Run migrations
	if err := config.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v\n", err)
	}

	// Load tax brackets from database
	repo := repositories.NewTaxBracketRepository(db)
	brackets, version, err := repo.GetActiveBrackets()
	if err != nil {
		log.Fatalf("Failed to load tax brackets from database: %v\n", err)
	}
	log.Printf("Loaded %d tax brackets (version %d) from database\n", len(brackets), version)

	// Cache brackets to GOB file
	if err := pkg.SaveToGob(TaxBracketsFile, brackets); err != nil {
		log.Printf("Warning: Failed to cache tax brackets to GOB file: %v\n", err)
	} else {
		log.Printf("Tax brackets cached to %s\n", TaxBracketsFile)
	}

	strategy := &domains.ProgressiveTaxStrategy{Brackets: brackets}
	service := services.NewPayslipService(strategy)
	employeeRepo := repositories.NewEmployeeRepository(db)

	// Run as CLI or Web based on APP_MODE
	if cfg.AppMode == "cli" {
		runCLI(service, employeeRepo)
	} else {
		runWeb(cfg, service, employeeRepo, repo, strategy)
	}

	// Cleanup GOB file on exit
	if err := os.Remove(TaxBracketsFile); err != nil {
		log.Printf("Warning: Failed to delete %s: %v\n", TaxBracketsFile, err)
	} else {
		log.Printf("Deleted %s\n", TaxBracketsFile)
	}
}

func runCLI(service *services.PayslipService, employeeRepo *repositories.EmployeeRepository) {
	log.Println("Running in CLI mode")
	cliHandler := &handlers.CLIHandler{Service: service, EmployeeRepo: employeeRepo}
	cliHandler.Run()
}

func runWeb(cfg *config.Config, service *services.PayslipService, employeeRepo *repositories.EmployeeRepository, taxBracketRepo *repositories.TaxBracketRepository, strategy *domains.ProgressiveTaxStrategy) {
	emailService := services.NewEmailService(services.SMTPConfig{
		APIKey:   cfg.EmailServerAPIKey,
		Host:     cfg.EmailSMTPHost,
		User:     cfg.EmailSMTPUser,
		Password: cfg.EmailSMTPPass,
		MailFrom: cfg.EmailSMTPMailFrom,
	})

	payslipHandler := &handlers.PayslipHandler{
		Service:        service,
		EmployeeRepo:   employeeRepo,
		EmailService:   emailService,
		TaxBracketRepo: taxBracketRepo,
		Strategy:       strategy,
	}

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())

	app.Post("/gen_monthly_payslip", payslipHandler.GenMonthlyPayslip)
	app.Get("/employees", payslipHandler.GetAllEmployees)
	app.Post("/send_email", payslipHandler.SendEmail)
	app.Post("/set_tax_brackets", payslipHandler.SetTaxBrackets)
	app.Post("/set_tax_brackets_active", payslipHandler.SetTaxBracketsActive)

	listenConfig := fiber.ListenConfig{
		EnablePrefork: cfg.Prefork,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		if err := app.Shutdown(); err != nil {
			log.Fatalf("Server shutdown failed: %v\n", err)
		}
		log.Println("Server shutdown complete")
	}()

	log.Printf("Server starting on port %s (env: %s, prefork: %v)", cfg.AppPort, cfg.AppEnv, cfg.Prefork)
	if err := app.Listen(":"+cfg.AppPort, listenConfig); err != nil {
		log.Printf("Server stopped: %v\n", err)
	}
}
