package handlers

import (
	"github.com/gofiber/fiber/v2"

	"be-zor/internal/store"
	"be-zor/internal/summary"
)

type AnalystHandler struct {
	store *store.BunStore
}

func NewAnalystHandler(bunStore *store.BunStore) *AnalystHandler {
	return &AnalystHandler{
		store: bunStore,
	}
}

func (h *AnalystHandler) Overview(c *fiber.Ctx) error {
	transactions, err := h.store.ListAllTransactions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load organization transactions",
		})
	}

	return c.JSON(fiber.Map{
		"summary":       summary.BuildOrganizationSummary(transactions),
		"recordsByUser": summary.BuildAnalystUserTransactions(transactions),
	})
}
