package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"be-zor/internal/models"
	"be-zor/internal/store"
)

func RequireAuth(bunStore *store.BunStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c.Get("Authorization"))
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization bearer token is required",
			})
		}

		sessionID := c.Get("X-Session-ID")
		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "x-session-id header is required",
			})
		}

		session, user, err := bunStore.ValidateSession(c.Context(), token, sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		c.Locals("auth.session", session)
		c.Locals("auth.user", user)

		return c.Next()
	}
}

func CurrentUser(c *fiber.Ctx) (models.User, bool) {
	user, ok := c.Locals("auth.user").(models.User)
	return user, ok
}

func CurrentSession(c *fiber.Ctx) (models.Session, bool) {
	session, ok := c.Locals("auth.session").(models.Session)
	return session, ok
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
