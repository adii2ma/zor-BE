package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"be-zor/internal/middleware"
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

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.store.ListManagedUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
	})
}

func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	input, err := parseAdminUserCreateRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user, err := h.store.CreateManagedUser(c.Context(), input)
	if err != nil {
		return adminUserStoreError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": user,
	})
}

func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("userID"))
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	input, err := parseAdminUserUpdateRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authenticated user is unavailable",
		})
	}
	if currentUser.ID == userID && input.Role != models.UserRoleAdmin {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "admin cannot remove their own admin role",
		})
	}
	if currentUser.ID == userID && input.Status != models.UserStatusActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "admin cannot deactivate their own account",
		})
	}

	user, err := h.store.UpdateManagedUser(c.Context(), userID, input)
	if err != nil {
		return adminUserStoreError(c, err)
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("userID"))
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authenticated user is unavailable",
		})
	}
	if currentUser.ID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "admin cannot delete their own account",
		})
	}

	if err := h.store.DeleteUser(c.Context(), userID); err != nil {
		return adminUserStoreError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) ListTransactions(c *fiber.Ctx) error {
	transactions, err := h.store.ListAllTransactions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load organization transactions",
		})
	}

	return c.JSON(fiber.Map{
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
		return adminTransactionStoreError(c, err)
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
		return adminTransactionStoreError(c, err)
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
		return adminTransactionStoreError(c, err)
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

func parseAdminUserCreateRequest(c *fiber.Ctx) (models.AdminUserCreateInput, error) {
	var request models.AdminUserCreateRequest
	if err := c.BodyParser(&request); err != nil {
		return models.AdminUserCreateInput{}, errors.New("request body is invalid")
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)
	request.Password = strings.TrimSpace(request.Password)

	if request.Name == "" {
		return models.AdminUserCreateInput{}, errors.New("name is required")
	}
	if request.Email == "" {
		return models.AdminUserCreateInput{}, errors.New("email is required")
	}
	if request.Password == "" {
		return models.AdminUserCreateInput{}, errors.New("password is required")
	}
	if len(request.Password) < 8 {
		return models.AdminUserCreateInput{}, errors.New("password must be at least 8 characters long")
	}
	if request.Role != models.UserRoleViewer && request.Role != models.UserRoleAnalyst && request.Role != models.UserRoleAdmin {
		return models.AdminUserCreateInput{}, errors.New("role must be viewer, analyst, or admin")
	}
	if request.Status != models.UserStatusActive && request.Status != models.UserStatusInactive {
		return models.AdminUserCreateInput{}, errors.New("status must be active or inactive")
	}

	return models.AdminUserCreateInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
		Status:   request.Status,
	}, nil
}

func parseAdminUserUpdateRequest(c *fiber.Ctx) (models.AdminUserUpdateInput, error) {
	var request models.AdminUserUpdateRequest
	if err := c.BodyParser(&request); err != nil {
		return models.AdminUserUpdateInput{}, errors.New("request body is invalid")
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)
	request.Password = strings.TrimSpace(request.Password)

	if request.Name == "" {
		return models.AdminUserUpdateInput{}, errors.New("name is required")
	}
	if request.Email == "" {
		return models.AdminUserUpdateInput{}, errors.New("email is required")
	}
	if request.Password != "" && len(request.Password) < 8 {
		return models.AdminUserUpdateInput{}, errors.New("password must be at least 8 characters long")
	}
	if request.Role != models.UserRoleViewer && request.Role != models.UserRoleAnalyst && request.Role != models.UserRoleAdmin {
		return models.AdminUserUpdateInput{}, errors.New("role must be viewer, analyst, or admin")
	}
	if request.Status != models.UserStatusActive && request.Status != models.UserStatusInactive {
		return models.AdminUserUpdateInput{}, errors.New("status must be active or inactive")
	}

	return models.AdminUserUpdateInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
		Status:   request.Status,
	}, nil
}

func adminUserStoreError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, store.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user was not found",
		})
	case errors.Is(err, store.ErrEmailAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "an account with this email already exists",
		})
	case errors.Is(err, store.ErrPasswordNotAllowed):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password updates are only allowed for local users",
		})
	case err.Error() == "email address is invalid":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user operation failed",
		})
	}
}

func adminTransactionStoreError(c *fiber.Ctx, err error) error {
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
