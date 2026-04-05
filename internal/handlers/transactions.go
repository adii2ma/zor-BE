package handlers

import (
	"github.com/gofiber/fiber/v2"

	"be-zor/internal/middleware"
	"be-zor/internal/store"
)

type TransactionHandler struct {
	store *store.BunStore
}

func NewTransactionHandler(bunStore *store.BunStore) *TransactionHandler {
	return &TransactionHandler{
		store: bunStore,
	}
}

func (h *TransactionHandler) List(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authenticated user is unavailable",
		})
	}

	transactions, err := h.store.ListTransactionsByUser(c.Context(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load user transactions",
		})
	}

	return c.JSON(fiber.Map{
		"user":         user,
		"transactions": transactions,
	})
}
