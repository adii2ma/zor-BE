package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"be-zor/internal/config"
	"be-zor/internal/googleauth"
	"be-zor/internal/handlers"
	"be-zor/internal/middleware"
	"be-zor/internal/store"
)

func main() {
	cfg := config.Load()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.FrontendOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Session-ID",
		AllowMethods: "GET,POST,OPTIONS",
	}))

	memoryStore := store.NewMemoryStore(cfg.SessionTTL)
	verifier := googleauth.NewVerifier(cfg.GoogleClientID)
	authHandler := handlers.NewAuthHandler(verifier, memoryStore)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "be-zor",
			"status":  "running",
		})
	})

	api := app.Group("/api")
	api.Post("/auth/google/signup", authHandler.GoogleSignup)
	api.Post("/auth/google/signin", authHandler.GoogleSignup)

	protected := api.Group("", middleware.RequireAuth(memoryStore))
	protected.Get("/me", authHandler.Me)

	log.Fatal(app.Listen(":" + cfg.Port))
}
