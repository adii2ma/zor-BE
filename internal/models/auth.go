package models

import "time"

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

type GoogleAuthRequest struct {
	Credential string `json:"credential"`
}

type AuthResponse struct {
	SessionToken string  `json:"sessionToken"`
	Session      Session `json:"session"`
	User         User    `json:"user"`
}
