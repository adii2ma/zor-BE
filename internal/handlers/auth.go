package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"be-zor/internal/googleauth"
	"be-zor/internal/models"
	"be-zor/internal/store"
)

type AuthHandler struct {
	verifier *googleauth.Verifier
	store    *store.MemoryStore
}

func NewAuthHandler(verifier *googleauth.Verifier, memoryStore *store.MemoryStore) *AuthHandler {
	return &AuthHandler{
		verifier: verifier,
		store:    memoryStore,
	}
}

func (h *AuthHandler) GoogleSignup(c *fiber.Ctx) error {
	var request models.GoogleAuthRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "request body is invalid",
		})
	}

	if request.Credential == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "google credential is required",
		})
	}

	identity, err := h.verifier.VerifyIDToken(c.Context(), request.Credential)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user := h.store.UpsertGoogleUser(identity)
	session, err := h.store.CreateSession(user.ID, c.Get("User-Agent"), c.IP())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
		SessionToken: session.Token,
		Session:      session,
		User:         user,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, ok := c.Locals("auth.user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authenticated user is unavailable",
		})
	}

	session, ok := c.Locals("auth.session").(models.Session)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authenticated session is unavailable",
		})
	}

	return c.JSON(fiber.Map{
		"user":    user,
		"session": session,
	})
}

func SessionErrorStatus(err error) int {
	switch {
	case errors.Is(err, store.ErrSessionNotFound):
		return fiber.StatusUnauthorized
	case errors.Is(err, store.ErrSessionExpired):
		return fiber.StatusUnauthorized
	case errors.Is(err, store.ErrUserNotFound):
		return fiber.StatusUnauthorized
	default:
		return fiber.StatusInternalServerError
	}
}
