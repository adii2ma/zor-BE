package store

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"be-zor/internal/models"
)

type adminUserRow struct {
	ID               string            `bun:"id"`
	Provider         string            `bun:"provider"`
	Role             models.UserRole   `bun:"role"`
	Status           models.UserStatus `bun:"status"`
	Email            string            `bun:"email"`
	Name             string            `bun:"name"`
	CreatedAt        time.Time         `bun:"created_at"`
	UpdatedAt        time.Time         `bun:"updated_at"`
	LastLoginAt      time.Time         `bun:"last_login_at"`
	TransactionCount int64             `bun:"transaction_count"`
}

func (s *BunStore) ListManagedUsers(
	ctx context.Context,
) ([]models.AdminUser, error) {
	var rows []adminUserRow
	if err := s.db.NewSelect().
		TableExpr("users AS u").
		ColumnExpr("u.id, u.provider, u.role, u.status, u.email, u.name, u.created_at, u.updated_at, u.last_login_at, COUNT(t.id) AS transaction_count").
		Join("LEFT JOIN transactions AS t ON t.user_id = u.id").
		GroupExpr("u.id, u.provider, u.role, u.status, u.email, u.name, u.created_at, u.updated_at, u.last_login_at").
		OrderExpr("u.created_at DESC, u.email ASC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	users := make([]models.AdminUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, models.AdminUser{
			ID:               row.ID,
			Provider:         row.Provider,
			Role:             row.Role,
			Status:           row.Status,
			Email:            row.Email,
			Name:             row.Name,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			LastLoginAt:      row.LastLoginAt,
			TransactionCount: int(row.TransactionCount),
		})
	}

	return users, nil
}

func (s *BunStore) CreateManagedUser(
	ctx context.Context,
	input models.AdminUserCreateInput,
) (models.User, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)

	if _, err := mail.ParseAddress(email); err != nil {
		return models.User{}, errors.New("email address is invalid")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	now := time.Now().UTC()
	record := models.NewLocalUserRecord(name, email, string(passwordHash), now)
	record.Role = input.Role
	record.Status = input.Status

	if _, err := s.db.NewInsert().Model(&record).Exec(ctx); err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailAlreadyExists
		}
		return models.User{}, err
	}

	return record.ToUser(), nil
}

func (s *BunStore) UpdateManagedUser(
	ctx context.Context,
	userID string,
	input models.AdminUserUpdateInput,
) (models.User, error) {
	var record models.UserRecord
	if err := s.db.NewSelect().
		Model(&record).
		Where("id = ?", userID).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	email := normalizeEmail(input.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return models.User{}, errors.New("email address is invalid")
	}

	record.Name = strings.TrimSpace(input.Name)
	record.Email = email
	record.Role = input.Role
	record.Status = input.Status
	record.UpdatedAt = time.Now().UTC()

	columns := []string{"name", "email", "role", "status", "updated_at"}
	if strings.TrimSpace(input.Password) != "" {
		if record.Provider != "local" {
			return models.User{}, ErrPasswordNotAllowed
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, err
		}
		record.PasswordHash = string(passwordHash)
		columns = append(columns, "password_hash")
	}

	if _, err := s.db.NewUpdate().
		Model(&record).
		WherePK().
		Column(columns...).
		Exec(ctx); err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrEmailAlreadyExists
		}
		return models.User{}, err
	}

	return record.ToUser(), nil
}

func (s *BunStore) DeleteUser(
	ctx context.Context,
	userID string,
) error {
	result, err := s.db.NewDelete().
		Model((*models.UserRecord)(nil)).
		Where("id = ?", userID).
		Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
