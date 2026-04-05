package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"be-zor/internal/models"
	"be-zor/internal/store"
)

type AdminHandler struct {
	store *store.BunStore
}

func NewAdminHandler(bunStore *store.BunStore) *AdminHandler {
	return &AdminHandler{
		store: bunStore,
	}
}

func (h *AdminHandler) ListTransactions(c *fiber.Ctx) error {
	transactions, err := h.store.ListAllTransactions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load organization transactions",
		})
	}

	users, err := h.store.ListUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load users",
		})
	}

	return c.JSON(fiber.Map{
		"users":        users,
		"transactions": transactions,
	})
}

func (h *AdminHandler) CreateTransaction(c *fiber.Ctx) error {
	input, err := parseTransactionMutationRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	transaction, err := h.store.CreateTransaction(c.Context(), input)
	if err != nil {
		return adminStoreError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"transaction": transaction,
	})
}

func (h *AdminHandler) UpdateTransaction(c *fiber.Ctx) error {
	transactionID := strings.TrimSpace(c.Params("transactionID"))
	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "transaction id is required",
		})
	}

	input, err := parseTransactionMutationRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	transaction, err := h.store.UpdateTransaction(c.Context(), transactionID, input)
	if err != nil {
		return adminStoreError(c, err)
	}

	return c.JSON(fiber.Map{
		"transaction": transaction,
	})
}

func (h *AdminHandler) DeleteTransaction(c *fiber.Ctx) error {
	transactionID := strings.TrimSpace(c.Params("transactionID"))
	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "transaction id is required",
		})
	}

	if err := h.store.DeleteTransaction(c.Context(), transactionID); err != nil {
		return adminStoreError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func parseTransactionMutationRequest(c *fiber.Ctx) (models.TransactionMutationInput, error) {
	var request models.TransactionMutationRequest
	if err := c.BodyParser(&request); err != nil {
		return models.TransactionMutationInput{}, errors.New("request body is invalid")
	}

	request.UserID = strings.TrimSpace(request.UserID)
	request.Category = strings.TrimSpace(request.Category)
	request.Description = strings.TrimSpace(request.Description)

	if request.UserID == "" {
		return models.TransactionMutationInput{}, errors.New("userId is required")
	}
	if request.Amount <= 0 {
		return models.TransactionMutationInput{}, errors.New("amount must be greater than zero")
	}
	if request.Type != models.TransactionTypeIncome && request.Type != models.TransactionTypeExpense {
		return models.TransactionMutationInput{}, errors.New("type must be income or expense")
	}
	if request.Category == "" {
		return models.TransactionMutationInput{}, errors.New("category is required")
	}
	if strings.TrimSpace(request.TransactionDate) == "" {
		return models.TransactionMutationInput{}, errors.New("transactionDate is required")
	}

	transactionDate, err := time.Parse("2006-01-02", request.TransactionDate)
	if err != nil {
		return models.TransactionMutationInput{}, errors.New("transactionDate must be in YYYY-MM-DD format")
	}

	return models.TransactionMutationInput{
		UserID:          request.UserID,
		Amount:          request.Amount,
		Type:            request.Type,
		Category:        request.Category,
		TransactionDate: transactionDate,
		Description:     request.Description,
	}, nil
}

func adminStoreError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, store.ErrUserNotFound):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "target user was not found",
		})
	case errors.Is(err, store.ErrTransactionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "transaction was not found",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "transaction operation failed",
		})
	}
}
