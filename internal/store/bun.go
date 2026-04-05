package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"be-zor/internal/models"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrUserNotFound    = errors.New("user not found")
)

type BunStore struct {
	db         *bun.DB
	sessionTTL time.Duration
}

func NewBunStore(db *bun.DB, sessionTTL time.Duration) *BunStore {
	return &BunStore{
		db:         db,
		sessionTTL: sessionTTL,
	}
}

func (s *BunStore) UpsertGoogleUser(
	ctx context.Context,
	identity models.GoogleIdentity,
) (models.User, bool, error) {
	now := time.Now().UTC()
	var record models.UserRecord

	err := s.db.NewSelect().
		Model(&record).
		Where("google_subject = ?", identity.Subject).
		Limit(1).
		Scan(ctx)

	switch {
	case err == nil:
		record.ApplyGoogleIdentity(identity, now)
		if _, err := s.db.NewUpdate().
			Model(&record).
			WherePK().
			Column(
				"provider",
				"google_subject",
				"email",
				"email_verified",
				"name",
				"given_name",
				"family_name",
				"picture",
				"locale",
				"hosted_domain",
				"google_issuer",
				"google_authorized_party",
				"google_audience",
				"google_issued_at",
				"google_expires_at",
				"updated_at",
				"last_login_at",
			).
			Exec(ctx); err != nil {
			return models.User{}, false, err
		}

		return record.ToUser(), false, nil
	case errors.Is(err, sql.ErrNoRows):
		record = models.NewUserRecord(identity, now)
		if _, err := s.db.NewInsert().Model(&record).Exec(ctx); err != nil {
			return models.User{}, false, err
		}

		return record.ToUser(), true, nil
	default:
		return models.User{}, false, err
	}
}

func (s *BunStore) CreateSession(
	ctx context.Context,
	userID string,
	userAgent string,
	ipAddress string,
) (models.Session, string, error) {
	var userRecord models.UserRecord
	if err := s.db.NewSelect().
		Model(&userRecord).
		Where("id = ?", userID).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Session{}, "", ErrUserNotFound
		}
		return models.Session{}, "", err
	}

	token, err := generateToken()
	if err != nil {
		return models.Session{}, "", err
	}

	now := time.Now().UTC()
	record := models.NewSessionRecord(
		userID,
		token,
		userAgent,
		ipAddress,
		now,
		now.Add(s.sessionTTL),
	)

	if _, err := s.db.NewInsert().Model(&record).Exec(ctx); err != nil {
		return models.Session{}, "", err
	}

	return record.ToSession(), token, nil
}

func (s *BunStore) ValidateSession(
	ctx context.Context,
	token string,
	sessionID string,
) (models.Session, models.User, error) {
	var sessionRecord models.SessionRecord
	if err := s.db.NewSelect().
		Model(&sessionRecord).
		Where("id = ?", sessionID).
		Where("token = ?", token).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Session{}, models.User{}, ErrSessionNotFound
		}
		return models.Session{}, models.User{}, err
	}

	now := time.Now().UTC()
	if now.After(sessionRecord.ExpiresAt) {
		_, _ = s.db.NewDelete().Model(&sessionRecord).WherePK().Exec(ctx)
		return models.Session{}, models.User{}, ErrSessionExpired
	}

	var userRecord models.UserRecord
	if err := s.db.NewSelect().
		Model(&userRecord).
		Where("id = ?", sessionRecord.UserID).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Session{}, models.User{}, ErrUserNotFound
		}
		return models.Session{}, models.User{}, err
	}

	sessionRecord.LastUsedAt = now
	if _, err := s.db.NewUpdate().
		Model(&sessionRecord).
		WherePK().
		Column("last_used_at").
		Exec(ctx); err != nil {
		return models.Session{}, models.User{}, err
	}

	return sessionRecord.ToSession(), userRecord.ToUser(), nil
}

func (s *BunStore) ListTransactionsByUser(
	ctx context.Context,
	userID string,
) ([]models.Transaction, error) {
	var records []models.TransactionRecord
	if err := s.db.NewSelect().
		Model(&records).
		Where("user_id = ?", userID).
		OrderExpr("transaction_date DESC, created_at DESC").
		Scan(ctx); err != nil {
		return nil, err
	}

	transactions := make([]models.Transaction, 0, len(records))
	for _, record := range records {
		transactions = append(transactions, record.ToTransaction())
	}

	return transactions, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
