package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"be-zor/internal/models"
)

type adminTransactionRow struct {
	ID              string                 `bun:"id"`
	UserID          string                 `bun:"user_id"`
	UserName        string                 `bun:"user_name"`
	UserEmail       string                 `bun:"user_email"`
	Amount          float64                `bun:"amount"`
	Type            models.TransactionType `bun:"type"`
	Category        string                 `bun:"category"`
	TransactionDate time.Time              `bun:"transaction_date"`
	Description     string                 `bun:"description"`
	CreatedAt       time.Time              `bun:"created_at"`
	UpdatedAt       time.Time              `bun:"updated_at"`
}

func (s *BunStore) ListAllTransactions(
	ctx context.Context,
) ([]models.AdminTransaction, error) {
	var rows []adminTransactionRow
	if err := s.db.NewSelect().
		TableExpr("transactions AS t").
		ColumnExpr("t.id, t.user_id, u.name AS user_name, u.email AS user_email, t.amount, t.type, t.category, t.transaction_date, t.description, t.created_at, t.updated_at").
		Join("JOIN users AS u ON u.id = t.user_id").
		OrderExpr("u.name ASC, t.transaction_date DESC, t.created_at DESC").
		Scan(ctx, &rows); err != nil {
		return nil, err
	}

	transactions := make([]models.AdminTransaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, models.AdminTransaction{
			ID:              row.ID,
			UserID:          row.UserID,
			UserName:        row.UserName,
			UserEmail:       row.UserEmail,
			Amount:          row.Amount,
			Type:            row.Type,
			Category:        row.Category,
			TransactionDate: row.TransactionDate,
			Description:     row.Description,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return transactions, nil
}

func (s *BunStore) ListUsers(
	ctx context.Context,
) ([]models.AdminUserOption, error) {
	var users []models.AdminUserOption
	if err := s.db.NewSelect().
		TableExpr("users AS u").
		ColumnExpr("u.id, u.name, u.email, u.role").
		OrderExpr("name ASC, email ASC").
		Scan(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *BunStore) CreateTransaction(
	ctx context.Context,
	input models.TransactionMutationInput,
) (models.Transaction, error) {
	if err := s.ensureUserExists(ctx, input.UserID); err != nil {
		return models.Transaction{}, err
	}

	now := time.Now().UTC()
	record := models.NewTransactionRecord(input, now)
	if _, err := s.db.NewInsert().Model(&record).Exec(ctx); err != nil {
		return models.Transaction{}, err
	}

	return record.ToTransaction(), nil
}

func (s *BunStore) UpdateTransaction(
	ctx context.Context,
	transactionID string,
	input models.TransactionMutationInput,
) (models.Transaction, error) {
	if err := s.ensureUserExists(ctx, input.UserID); err != nil {
		return models.Transaction{}, err
	}

	var record models.TransactionRecord
	if err := s.db.NewSelect().
		Model(&record).
		Where("id = ?", transactionID).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Transaction{}, ErrTransactionNotFound
		}
		return models.Transaction{}, err
	}

	record.UserID = input.UserID
	record.Amount = input.Amount
	record.Type = input.Type
	record.Category = input.Category
	record.TransactionDate = input.TransactionDate.UTC()
	record.Description = input.Description
	record.UpdatedAt = time.Now().UTC()

	if _, err := s.db.NewUpdate().
		Model(&record).
		WherePK().
		Column("user_id", "amount", "type", "category", "transaction_date", "description", "updated_at").
		Exec(ctx); err != nil {
		return models.Transaction{}, err
	}

	return record.ToTransaction(), nil
}

func (s *BunStore) DeleteTransaction(
	ctx context.Context,
	transactionID string,
) error {
	result, err := s.db.NewDelete().
		Model((*models.TransactionRecord)(nil)).
		Where("id = ?", transactionID).
		Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

func (s *BunStore) ensureUserExists(ctx context.Context, userID string) error {
	var user models.UserRecord
	if err := s.db.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
