package store

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"be-zor/internal/models"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrUserNotFound    = errors.New("user not found")
)

type MemoryStore struct {
	mu               sync.RWMutex
	sessionTTL       time.Duration
	usersByID        map[string]*models.User
	usersByGoogleSub map[string]string
	sessionsByToken  map[string]*models.Session
}

func NewMemoryStore(sessionTTL time.Duration) *MemoryStore {
	return &MemoryStore{
		sessionTTL:       sessionTTL,
		usersByID:        make(map[string]*models.User),
		usersByGoogleSub: make(map[string]string),
		sessionsByToken:  make(map[string]*models.Session),
	}
}

func (s *MemoryStore) UpsertGoogleUser(identity models.GoogleIdentity) models.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if userID, ok := s.usersByGoogleSub[identity.Subject]; ok {
		existing := s.usersByID[userID]
		existing.Email = identity.Email
		existing.EmailVerified = identity.EmailVerified
		existing.Name = identity.Name
		existing.GivenName = identity.GivenName
		existing.FamilyName = identity.FamilyName
		existing.Picture = identity.Picture
		existing.Locale = identity.Locale
		existing.HostedDomain = identity.HostedDomain
		existing.Google = identity
		existing.UpdatedAt = now
		existing.LastLoginAt = now
		return *existing
	}

	user := &models.User{
		ID:            uuid.NewString(),
		Provider:      "google",
		Email:         identity.Email,
		EmailVerified: identity.EmailVerified,
		Name:          identity.Name,
		GivenName:     identity.GivenName,
		FamilyName:    identity.FamilyName,
		Picture:       identity.Picture,
		Locale:        identity.Locale,
		HostedDomain:  identity.HostedDomain,
		Google:        identity,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   now,
	}

	s.usersByID[user.ID] = user
	s.usersByGoogleSub[identity.Subject] = user.ID

	return *user
}

func (s *MemoryStore) CreateSession(userID, userAgent, ipAddress string) (models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.usersByID[userID]
	if user == nil {
		return models.Session{}, ErrUserNotFound
	}

	token, err := generateToken()
	if err != nil {
		return models.Session{}, err
	}

	now := time.Now().UTC()
	session := &models.Session{
		ID:         uuid.NewString(),
		UserID:     userID,
		Token:      token,
		Provider:   "google",
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.sessionTTL),
		LastUsedAt: now,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}

	s.sessionsByToken[token] = session
	return *session, nil
}

func (s *MemoryStore) ValidateSession(token, sessionID string) (models.Session, models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessionsByToken[token]
	if session == nil || session.ID != sessionID {
		return models.Session{}, models.User{}, ErrSessionNotFound
	}

	now := time.Now().UTC()
	if now.After(session.ExpiresAt) {
		delete(s.sessionsByToken, token)
		return models.Session{}, models.User{}, ErrSessionExpired
	}

	user := s.usersByID[session.UserID]
	if user == nil {
		delete(s.sessionsByToken, token)
		return models.Session{}, models.User{}, ErrUserNotFound
	}

	session.LastUsedAt = now
	return *session, *user, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
