package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"be-zor/internal/config"
	"be-zor/internal/database"
	"be-zor/internal/googleauth"
	"be-zor/internal/handlers"
	"be-zor/internal/middleware"
	"be-zor/internal/models"
	"be-zor/internal/store"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if cfg.ShouldMigrate {
		group, err := database.Migrate(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("migration status: %s", database.FormatMigrationGroup(group))
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.FrontendOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Session-ID",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
	}))

	bunStore := store.NewBunStore(db, cfg.SessionTTL)
	verifier := googleauth.NewVerifier(cfg.GoogleClientID)
	authHandler := handlers.NewAuthHandler(verifier, bunStore)
	dashboardHandler := handlers.NewDashboardHandler(bunStore)
	transactionHandler := handlers.NewTransactionHandler(bunStore)
	analystHandler := handlers.NewAnalystHandler(bunStore)
	adminHandler := handlers.NewAdminHandler(bunStore)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "be-zor",
			"status":  "running",
		})
	})

	api := app.Group("/api")
	api.Post("/auth/google/signup", authHandler.GoogleSignup)
	api.Post("/auth/google/signin", authHandler.GoogleSignup)
	api.Post("/auth/local/signup", authHandler.LocalSignUp)
	api.Post("/auth/local/signin", authHandler.LocalSignIn)

	protected := api.Group("", middleware.RequireAuth(bunStore))
	protected.Get("/me", authHandler.Me)
	protected.Get("/dashboard/summary", dashboardHandler.Summary)
	protected.Get("/transactions", transactionHandler.List)

	analyst := protected.Group("/analyst", middleware.RequireRole(models.UserRoleAnalyst, models.UserRoleAdmin))
	analyst.Get("/overview", analystHandler.Overview)

	admin := protected.Group("/admin", middleware.RequireRole(models.UserRoleAdmin))
	admin.Get("/transactions", adminHandler.ListTransactions)
	admin.Post("/transactions", adminHandler.CreateTransaction)
	admin.Patch("/transactions/:transactionID", adminHandler.UpdateTransaction)
	admin.Delete("/transactions/:transactionID", adminHandler.DeleteTransaction)

	log.Fatal(app.Listen(":" + cfg.Port))
}
