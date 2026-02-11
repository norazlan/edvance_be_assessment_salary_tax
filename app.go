package main

import (
	"log"

	"edvance-assessment/config"
	"edvance-assessment/internal/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())

	payslipHandler := handlers.NewPayslipHandler()

	app.Post("/gen_monthly_payslip", payslipHandler.GenMonthlyPayslip)

	listenConfig := fiber.ListenConfig{
		EnablePrefork: cfg.Prefork,
	}

	log.Printf("Server starting on port %s (env: %s, prefork: %v)", cfg.AppPort, cfg.AppEnv, cfg.Prefork)
	log.Fatal(app.Listen(":"+cfg.AppPort, listenConfig))
}
