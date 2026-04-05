package handlers

import (
	"errors"
	"log"

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

	log.Printf("google signup hit: email=%s subject=%s", identity.Email, identity.Subject)

	user, created, err := h.store.UpsertGoogleUser(c.Context(), identity)
	if err != nil {
		log.Printf("google signup db error: email=%s err=%v", identity.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to persist user",
		})
	}

	if created {
		log.Printf("google signup user created in db: email=%s user_id=%s", user.Email, user.ID)
	} else {
		log.Printf("google signup user already existed in db: email=%s user_id=%s", user.Email, user.ID)
	}

	session, sessionToken, err := h.store.CreateSession(
		c.Context(),
		user.ID,
		c.Get("User-Agent"),
		c.IP(),
	)
	if err != nil {
		if errors.Is(err, store.ErrUserInactive) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "user account is inactive",
			})
		}
		log.Printf("google signup session create error: email=%s err=%v", user.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	log.Printf("google signup session created: email=%s session_id=%s", user.Email, session.ID)

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
