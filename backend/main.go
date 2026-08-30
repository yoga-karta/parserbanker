package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	InitAuth() // fail-fast kalau CLIENT_USER/CLIENT_PASS belum di-set

	if err := os.MkdirAll(tmpRoot, 0700); err != nil {
		log.Fatalf("gagal membuat tmp root: %v", err)
	}

	store := NewJobStore()

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, X-API-Token",
	}))

	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 60 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Terlalu banyak percobaan login. Silakan tunggu 1 menit.",
			})
		},
	})

	api := app.Group("/api")
	api.Post("/login", loginLimiter, handleLogin)

	api.Use(AuthMiddleware())

	// Flow 2 langkah: create -> load (ej & rc terpisah) -> process -> (stop opsional) -> results/export
	api.Post("/jobs", handleCreateJob(store))
	api.Post("/jobs/:id/load/:role", handleLoadFile(store))
	api.Post("/jobs/:id/process", handleProcess(store))
	api.Post("/jobs/:id/stop", handleStop(store))
	api.Post("/jobs/:id/reset", handleResetJob(store))
	api.Get("/jobs/:id", handleJobStatus(store))
	api.Get("/jobs/:id/log", handleJobLog(store))
	api.Get("/jobs/:id/results", handleJobResults(store))
	api.Get("/jobs/:id/export", handleExport(store))

	api.Post("/logout", handleLogout)
	api.Get("/history", handleHistory)
	api.Delete("/history", handleClearHistory)

	log.Fatal(app.Listen("127.0.0.1:8080"))
}
