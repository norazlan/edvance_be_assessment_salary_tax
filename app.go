package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"edvance-assessment/config"
	"edvance-assessment/internal/domains"
	"edvance-assessment/internal/handlers"
	"edvance-assessment/pkg"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

const TaxBracketsFile = "data/tax_brackets.gob"

func init() {
	brackets := []domains.TaxBracket{
		{Min: 0, Max: 20000, Rate: 0.0},
		{Min: 20001, Max: 40000, Rate: 0.1},
		{Min: 40001, Max: 80000, Rate: 0.2},
		{Min: 80001, Max: 180000, Rate: 0.3},
		{Min: 180001, Max: 999999999, Rate: 0.4},
	}

	if err := pkg.SaveToGob(TaxBracketsFile, brackets); err != nil {
		log.Fatalf("Warning: Failed to save tax brackets to GOB file: %v\n", err)
	} else {
		log.Printf("Tax brackets saved to %s\n", TaxBracketsFile)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())

	payslipHandler := handlers.NewPayslipHandler()

	app.Post("/gen_monthly_payslip", payslipHandler.GenMonthlyPayslip)

	listenConfig := fiber.ListenConfig{
		EnablePrefork: cfg.Prefork,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		// Delete TaxBracketsFile before shutdown
		if err := os.Remove(TaxBracketsFile); err != nil {
			log.Printf("Warning: Failed to delete %s: %v\n", TaxBracketsFile, err)
		} else {
			log.Printf("Deleted %s\n", TaxBracketsFile)
		}

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
