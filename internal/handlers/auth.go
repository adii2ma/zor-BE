package handlers

import (
	"github.com/gofiber/fiber/v2"

	"be-zor/internal/googleauth"
	"be-zor/internal/models"
	"be-zor/internal/store"
)

type AuthHandler struct {
	verifier *googleauth.Verifier
	store    *store.BunStore
}

func NewAuthHandler(verifier *googleauth.Verifier, bunStore *store.BunStore) *AuthHandler {
	return &AuthHandler{
		verifier: verifier,
		store:    bunStore,
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

	user, err := h.store.UpsertGoogleUser(c.Context(), identity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to persist user",
		})
	}

	session, sessionToken, err := h.store.CreateSession(
		c.Context(),
		user.ID,
		c.Get("User-Agent"),
		c.IP(),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
		SessionToken: sessionToken,
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
