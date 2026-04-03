package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type GoogleIdentity struct {
	Issuer          string    `json:"issuer"`
	AuthorizedParty string    `json:"authorizedParty,omitempty"`
	Audience        string    `json:"audience"`
	Subject         string    `json:"subject"`
	Email           string    `json:"email"`
	EmailVerified   bool      `json:"emailVerified"`
	Name            string    `json:"name"`
	GivenName       string    `json:"givenName,omitempty"`
	FamilyName      string    `json:"familyName,omitempty"`
	Picture         string    `json:"picture,omitempty"`
	Locale          string    `json:"locale,omitempty"`
	HostedDomain    string    `json:"hostedDomain,omitempty"`
	IssuedAt        time.Time `json:"issuedAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
}

type User struct {
	ID            string         `json:"id"`
	Provider      string         `json:"provider"`
	Email         string         `json:"email"`
	EmailVerified bool           `json:"emailVerified"`
	Name          string         `json:"name"`
	GivenName     string         `json:"givenName,omitempty"`
	FamilyName    string         `json:"familyName,omitempty"`
	Picture       string         `json:"picture,omitempty"`
	Locale        string         `json:"locale,omitempty"`
	HostedDomain  string         `json:"hostedDomain,omitempty"`
	Google        GoogleIdentity `json:"google"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	LastLoginAt   time.Time      `json:"lastLoginAt"`
}

type UserRecord struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                    string    `bun:"id,pk" json:"-"`
	Provider              string    `bun:"provider,notnull" json:"-"`
	GoogleSubject         string    `bun:"google_subject,notnull" json:"-"`
	Email                 string    `bun:"email,notnull" json:"-"`
	EmailVerified         bool      `bun:"email_verified,notnull" json:"-"`
	Name                  string    `bun:"name,notnull" json:"-"`
	GivenName             string    `bun:"given_name" json:"-"`
	FamilyName            string    `bun:"family_name" json:"-"`
	Picture               string    `bun:"picture" json:"-"`
	Locale                string    `bun:"locale" json:"-"`
	HostedDomain          string    `bun:"hosted_domain" json:"-"`
	GoogleIssuer          string    `bun:"google_issuer,notnull" json:"-"`
	GoogleAuthorizedParty string    `bun:"google_authorized_party" json:"-"`
	GoogleAudience        string    `bun:"google_audience,notnull" json:"-"`
	GoogleIssuedAt        time.Time `bun:"google_issued_at,notnull" json:"-"`
	GoogleExpiresAt       time.Time `bun:"google_expires_at,notnull" json:"-"`
	CreatedAt             time.Time `bun:"created_at,notnull" json:"-"`
	UpdatedAt             time.Time `bun:"updated_at,notnull" json:"-"`
	LastLoginAt           time.Time `bun:"last_login_at,notnull" json:"-"`
}

type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Token      string    `json:"-"`
	Provider   string    `json:"provider"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	UserAgent  string    `json:"userAgent,omitempty"`
	IPAddress  string    `json:"ipAddress,omitempty"`
}

type SessionRecord struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID         string    `bun:"id,pk" json:"-"`
	UserID     string    `bun:"user_id,notnull" json:"-"`
	Token      string    `bun:"token,notnull" json:"-"`
	Provider   string    `bun:"provider,notnull" json:"-"`
	CreatedAt  time.Time `bun:"created_at,notnull" json:"-"`
	ExpiresAt  time.Time `bun:"expires_at,notnull" json:"-"`
	LastUsedAt time.Time `bun:"last_used_at,notnull" json:"-"`
	UserAgent  string    `bun:"user_agent" json:"-"`
	IPAddress  string    `bun:"ip_address" json:"-"`
}

type GoogleAuthRequest struct {
	Credential string `json:"credential"`
}

type AuthResponse struct {
	SessionToken string  `json:"sessionToken"`
	Session      Session `json:"session"`
	User         User    `json:"user"`
}

func NewUserRecord(identity GoogleIdentity, now time.Time) UserRecord {
	record := UserRecord{
		ID:        uuid.NewString(),
		Provider:  "google",
		CreatedAt: now,
	}
	record.ApplyGoogleIdentity(identity, now)
	return record
}

func (u *UserRecord) ApplyGoogleIdentity(identity GoogleIdentity, now time.Time) {
	u.Provider = "google"
	u.GoogleSubject = identity.Subject
	u.Email = identity.Email
	u.EmailVerified = identity.EmailVerified
	u.Name = identity.Name
	u.GivenName = identity.GivenName
	u.FamilyName = identity.FamilyName
	u.Picture = identity.Picture
	u.Locale = identity.Locale
	u.HostedDomain = identity.HostedDomain
	u.GoogleIssuer = identity.Issuer
	u.GoogleAuthorizedParty = identity.AuthorizedParty
	u.GoogleAudience = identity.Audience
	u.GoogleIssuedAt = identity.IssuedAt
	u.GoogleExpiresAt = identity.ExpiresAt
	u.UpdatedAt = now
	u.LastLoginAt = now
}

func (u UserRecord) ToUser() User {
	return User{
		ID:            u.ID,
		Provider:      u.Provider,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Name:          u.Name,
		GivenName:     u.GivenName,
		FamilyName:    u.FamilyName,
		Picture:       u.Picture,
		Locale:        u.Locale,
		HostedDomain:  u.HostedDomain,
		Google: GoogleIdentity{
			Issuer:          u.GoogleIssuer,
			AuthorizedParty: u.GoogleAuthorizedParty,
			Audience:        u.GoogleAudience,
			Subject:         u.GoogleSubject,
			Email:           u.Email,
			EmailVerified:   u.EmailVerified,
			Name:            u.Name,
			GivenName:       u.GivenName,
			FamilyName:      u.FamilyName,
			Picture:         u.Picture,
			Locale:          u.Locale,
			HostedDomain:    u.HostedDomain,
			IssuedAt:        u.GoogleIssuedAt,
			ExpiresAt:       u.GoogleExpiresAt,
		},
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

func NewSessionRecord(
	userID string,
	token string,
	userAgent string,
	ipAddress string,
	now time.Time,
	expiresAt time.Time,
) SessionRecord {
	return SessionRecord{
		ID:         uuid.NewString(),
		UserID:     userID,
		Token:      token,
		Provider:   "google",
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastUsedAt: now,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}
}

func (s SessionRecord) ToSession() Session {
	return Session{
		ID:         s.ID,
		UserID:     s.UserID,
		Provider:   s.Provider,
		CreatedAt:  s.CreatedAt,
		ExpiresAt:  s.ExpiresAt,
		LastUsedAt: s.LastUsedAt,
		UserAgent:  s.UserAgent,
		IPAddress:  s.IPAddress,
	}
}
