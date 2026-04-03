package googleauth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"be-zor/internal/models"
)

const certificatesURL = "https://www.googleapis.com/oauth2/v1/certs"

type Verifier struct {
	clientID   string
	httpClient *http.Client

	mu              sync.Mutex
	certificates    map[string]*rsa.PublicKey
	certificatesTTL time.Time
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Type      string `json:"typ"`
}

type rawClaims struct {
	Issuer          string          `json:"iss"`
	AuthorizedParty string          `json:"azp"`
	Audience        string          `json:"aud"`
	Subject         string          `json:"sub"`
	Email           string          `json:"email"`
	EmailVerified   json.RawMessage `json:"email_verified"`
	Name            string          `json:"name"`
	GivenName       string          `json:"given_name"`
	FamilyName      string          `json:"family_name"`
	Picture         string          `json:"picture"`
	Locale          string          `json:"locale"`
	HostedDomain    string          `json:"hd"`
	IssuedAt        int64           `json:"iat"`
	ExpiresAt       int64           `json:"exp"`
}

func NewVerifier(clientID string) *Verifier {
	return &Verifier{
		clientID: clientID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (v *Verifier) VerifyIDToken(ctx context.Context, token string) (models.GoogleIdentity, error) {
	if v.clientID == "" {
		return models.GoogleIdentity{}, errors.New("google client id is not configured")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return models.GoogleIdentity{}, errors.New("google credential is not a valid jwt")
	}

	headerBytes, err := decodeJWTPart(parts[0])
	if err != nil {
		return models.GoogleIdentity{}, fmt.Errorf("decode jwt header: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return models.GoogleIdentity{}, fmt.Errorf("parse jwt header: %w", err)
	}

	if header.Algorithm != "RS256" || header.KeyID == "" {
		return models.GoogleIdentity{}, errors.New("google credential uses an unsupported signature")
	}

	certificates, err := v.getCertificates(ctx)
	if err != nil {
		return models.GoogleIdentity{}, err
	}

	publicKey, ok := certificates[header.KeyID]
	if !ok {
		return models.GoogleIdentity{}, errors.New("google signing key not found")
	}

	signature, err := decodeJWTPart(parts[2])
	if err != nil {
		return models.GoogleIdentity{}, fmt.Errorf("decode jwt signature: %w", err)
	}

	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, sum[:], signature); err != nil {
		return models.GoogleIdentity{}, errors.New("google credential signature verification failed")
	}

	payloadBytes, err := decodeJWTPart(parts[1])
	if err != nil {
		return models.GoogleIdentity{}, fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims rawClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return models.GoogleIdentity{}, fmt.Errorf("parse jwt claims: %w", err)
	}

	emailVerified, err := parseBoolOrString(claims.EmailVerified)
	if err != nil {
		return models.GoogleIdentity{}, err
	}

	if claims.Audience != v.clientID {
		return models.GoogleIdentity{}, errors.New("google credential audience mismatch")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return models.GoogleIdentity{}, errors.New("google credential issuer is invalid")
	}

	now := time.Now().UTC()
	expiresAt := time.Unix(claims.ExpiresAt, 0).UTC()
	if now.After(expiresAt) {
		return models.GoogleIdentity{}, errors.New("google credential is expired")
	}

	issuedAt := time.Unix(claims.IssuedAt, 0).UTC()
	if issuedAt.After(now.Add(2 * time.Minute)) {
		return models.GoogleIdentity{}, errors.New("google credential issue time is invalid")
	}

	return models.GoogleIdentity{
		Issuer:          claims.Issuer,
		AuthorizedParty: claims.AuthorizedParty,
		Audience:        claims.Audience,
		Subject:         claims.Subject,
		Email:           claims.Email,
		EmailVerified:   emailVerified,
		Name:            claims.Name,
		GivenName:       claims.GivenName,
		FamilyName:      claims.FamilyName,
		Picture:         claims.Picture,
		Locale:          claims.Locale,
		HostedDomain:    claims.HostedDomain,
		IssuedAt:        issuedAt,
		ExpiresAt:       expiresAt,
	}, nil
}

func (v *Verifier) getCertificates(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now().UTC()
	if len(v.certificates) > 0 && now.Before(v.certificatesTTL) {
		return v.certificates, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, certificatesURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := v.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch google certificates: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch google certificates: unexpected status %d", response.StatusCode)
	}

	var rawCertificates map[string]string
	if err := json.NewDecoder(response.Body).Decode(&rawCertificates); err != nil {
		return nil, fmt.Errorf("decode google certificates: %w", err)
	}

	parsedCertificates := make(map[string]*rsa.PublicKey, len(rawCertificates))
	for keyID, certificatePEM := range rawCertificates {
		publicKey, err := parseCertificate(certificatePEM)
		if err != nil {
			return nil, fmt.Errorf("parse google certificate %s: %w", keyID, err)
		}

		parsedCertificates[keyID] = publicKey
	}

	v.certificates = parsedCertificates
	v.certificatesTTL = now.Add(parseCacheDuration(response.Header.Get("Cache-Control")))

	return v.certificates, nil
}

func decodeJWTPart(part string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(part)
}

func parseCertificate(certificatePEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return nil, errors.New("certificate pem is invalid")
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("certificate public key is not rsa")
	}

	return publicKey, nil
}

func parseCacheDuration(cacheControl string) time.Duration {
	for _, directive := range strings.Split(cacheControl, ",") {
		directive = strings.TrimSpace(directive)
		if !strings.HasPrefix(directive, "max-age=") {
			continue
		}

		seconds, err := strconv.Atoi(strings.TrimPrefix(directive, "max-age="))
		if err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}

	return time.Hour
}

func parseBoolOrString(value json.RawMessage) (bool, error) {
	if len(value) == 0 {
		return false, nil
	}

	var boolValue bool
	if err := json.Unmarshal(value, &boolValue); err == nil {
		return boolValue, nil
	}

	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err == nil {
		return strings.EqualFold(stringValue, "true"), nil
	}

	return false, errors.New("google credential email verification flag is invalid")
}
