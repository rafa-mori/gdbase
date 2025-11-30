package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	gl "github.com/kubex-ecosystem/logz"
)

// IAuthCodeService interface for authorization code service operations
type IAuthCodeService interface {
	GenerateCode(userID, clientID, redirectURI, codeChallenge, method string, scope string, expirationMinutes int) (IAuthCode, error)
	GetByCode(code string) (IAuthCode, error)
	ValidateAndConsume(code string) (IAuthCode, error)
	GetUserCodes(userID string) ([]IAuthCode, error)
	CleanupExpired() error
}

// AuthCodeService implements IAuthCodeService
type AuthCodeService struct {
	repo IAuthCodeRepo
}

// NewAuthCodeService creates a new authorization code service
func NewAuthCodeService(repo IAuthCodeRepo) IAuthCodeService {
	if repo == nil {
		gl.Log("error", "AuthCodeService: repository is nil")
		return nil
	}
	return &AuthCodeService{repo: repo}
}

func (s *AuthCodeService) GenerateCode(userID, clientID, redirectURI, codeChallenge, method string, scope string, expirationMinutes int) (IAuthCode, error) {
	if userID == "" {
		return nil, fmt.Errorf("AuthCodeService: user_id is required")
	}
	if clientID == "" {
		return nil, fmt.Errorf("AuthCodeService: client_id is required")
	}
	if redirectURI == "" {
		return nil, fmt.Errorf("AuthCodeService: redirect_uri is required")
	}
	if codeChallenge == "" {
		return nil, fmt.Errorf("AuthCodeService: code_challenge is required")
	}
	if method != "S256" && method != "plain" {
		return nil, fmt.Errorf("AuthCodeService: invalid code_challenge_method (must be 'S256' or 'plain')")
	}

	if expirationMinutes <= 0 {
		expirationMinutes = 10 // Default: 10 minutes
	}

	// Generate cryptographically secure random code
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, fmt.Errorf("AuthCodeService: failed to generate random code: %w", err)
	}
	code := base64.RawURLEncoding.EncodeToString(codeBytes)

	authCode := &AuthCodeModel{
		Code:                code,
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: method,
		Scope:               scope,
		ExpiresAt:           time.Now().Add(time.Duration(expirationMinutes) * time.Minute),
		Used:                false,
	}

	if err := authCode.Validate(); err != nil {
		return nil, fmt.Errorf("AuthCodeService: validation failed: %w", err)
	}

	created, err := s.repo.Create(authCode)
	if err != nil {
		return nil, fmt.Errorf("AuthCodeService: failed to create code: %w", err)
	}

	gl.Log("info", fmt.Sprintf("AuthCodeService: generated code for user %s, client %s", userID, clientID))
	return created, nil
}

func (s *AuthCodeService) GetByCode(code string) (IAuthCode, error) {
	if code == "" {
		return nil, fmt.Errorf("AuthCodeService: code is required")
	}

	authCode, err := s.repo.FindByCode(code)
	if err != nil {
		return nil, fmt.Errorf("AuthCodeService: code not found: %w", err)
	}

	return authCode, nil
}

func (s *AuthCodeService) ValidateAndConsume(code string) (IAuthCode, error) {
	if code == "" {
		return nil, fmt.Errorf("AuthCodeService: code is required")
	}

	authCode, err := s.repo.FindByCode(code)
	if err != nil {
		return nil, fmt.Errorf("AuthCodeService: code not found: %w", err)
	}

	if authCode.GetUsed() {
		return nil, fmt.Errorf("AuthCodeService: code already used")
	}

	if authCode.IsExpired() {
		return nil, fmt.Errorf("AuthCodeService: code expired")
	}

	if err := s.repo.MarkAsUsed(code); err != nil {
		return nil, fmt.Errorf("AuthCodeService: failed to mark code as used: %w", err)
	}

	gl.Log("info", fmt.Sprintf("AuthCodeService: consumed code for user %s", authCode.GetUserID()))
	return authCode, nil
}

func (s *AuthCodeService) GetUserCodes(userID string) ([]IAuthCode, error) {
	if userID == "" {
		return nil, fmt.Errorf("AuthCodeService: user_id is required")
	}

	codes, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("AuthCodeService: failed to get user codes: %w", err)
	}

	return codes, nil
}

func (s *AuthCodeService) CleanupExpired() error {
	if err := s.repo.DeleteExpired(); err != nil {
		return fmt.Errorf("AuthCodeService: failed to cleanup expired codes: %w", err)
	}

	gl.Log("info", "AuthCodeService: cleaned up expired codes")
	return nil
}
