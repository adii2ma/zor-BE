package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"be-zor/internal/models"
	"be-zor/internal/store"
)

func (h *AuthHandler) LocalSignUp(c *fiber.Ctx) error {
	var request models.LocalSignUpRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "request body is invalid",
		})
	}

	name := strings.TrimSpace(request.Name)
	email := strings.TrimSpace(request.Email)

	switch {
	case name == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	case email == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	case request.Password == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password is required",
		})
	case len(request.Password) < 8:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 8 characters long",
		})
	case request.ConfirmPassword == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "confirm password is required",
		})
	case request.Password != request.ConfirmPassword:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password and confirm password must match",
		})
	}

	user, err := h.store.CreateLocalUser(c.Context(), name, email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrEmailAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "an account with this email already exists",
			})
		case err.Error() == "email address is invalid":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create account",
			})
		}
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

func (h *AuthHandler) LocalSignIn(c *fiber.Ctx) error {
	var request models.LocalSignInRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "request body is invalid",
		})
	}

	email := strings.TrimSpace(request.Email)

	switch {
	case email == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	case request.Password == "":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password is required",
		})
	}

	user, err := h.store.AuthenticateLocalUser(c.Context(), email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidCredentials):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "email or password is incorrect",
			})
		case errors.Is(err, store.ErrUserInactive):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "user account is inactive",
			})
		case err.Error() == "email address is invalid":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to sign in",
			})
		}
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.AuthResponse{
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	})
}
