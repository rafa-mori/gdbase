package oauth

import (
	"fmt"

	gl "github.com/kubex-ecosystem/logz"
)

// IOAuthClientService interface for OAuth client service operations
type IOAuthClientService interface {
	CreateClient(client IOAuthClient) (IOAuthClient, error)
	GetClientByID(id string) (IOAuthClient, error)
	GetClientByClientID(clientID string) (IOAuthClient, error)
	ListClients() ([]IOAuthClient, error)
	ListActiveClients() ([]IOAuthClient, error)
	UpdateClient(client IOAuthClient) (IOAuthClient, error)
	DeactivateClient(id string) error
	DeleteClient(id string) error
	ValidateRedirectURI(clientID, redirectURI string) error
}

// OAuthClientService implements IOAuthClientService
type OAuthClientService struct {
	repo IOAuthClientRepo
}

// NewOAuthClientService creates a new OAuth client service
func NewOAuthClientService(repo IOAuthClientRepo) IOAuthClientService {
	if repo == nil {
		gl.Log("error", "OAuthClientService: repository is nil")
		return nil
	}
	return &OAuthClientService{repo: repo}
}

func (s *OAuthClientService) CreateClient(client IOAuthClient) (IOAuthClient, error) {
	if client == nil {
		return nil, fmt.Errorf("OAuthClientService: client is nil")
	}

	if err := client.Validate(); err != nil {
		return nil, fmt.Errorf("OAuthClientService: validation failed: %w", err)
	}

	created, err := s.repo.Create(client)
	if err != nil {
		return nil, fmt.Errorf("OAuthClientService: failed to create client: %w", err)
	}

	gl.Log("info", fmt.Sprintf("OAuthClientService: created client %s (%s)", created.GetClientID(), created.GetClientName()))
	return created, nil
}

func (s *OAuthClientService) GetClientByID(id string) (IOAuthClient, error) {
	if id == "" {
		return nil, fmt.Errorf("OAuthClientService: id is required")
	}

	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("OAuthClientService: failed to get client: %w", err)
	}

	return client, nil
}

func (s *OAuthClientService) GetClientByClientID(clientID string) (IOAuthClient, error) {
	if clientID == "" {
		return nil, fmt.Errorf("OAuthClientService: client_id is required")
	}

	client, err := s.repo.FindByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("OAuthClientService: failed to get client: %w", err)
	}

	if !client.GetActive() {
		return nil, fmt.Errorf("OAuthClientService: client is inactive")
	}

	return client, nil
}

func (s *OAuthClientService) ListClients() ([]IOAuthClient, error) {
	clients, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("OAuthClientService: failed to list clients: %w", err)
	}

	return clients, nil
}

func (s *OAuthClientService) ListActiveClients() ([]IOAuthClient, error) {
	clients, err := s.repo.FindAll("active = ?", true)
	if err != nil {
		return nil, fmt.Errorf("OAuthClientService: failed to list active clients: %w", err)
	}

	return clients, nil
}

func (s *OAuthClientService) UpdateClient(client IOAuthClient) (IOAuthClient, error) {
	if client == nil {
		return nil, fmt.Errorf("OAuthClientService: client is nil")
	}

	if err := client.Validate(); err != nil {
		return nil, fmt.Errorf("OAuthClientService: validation failed: %w", err)
	}

	updated, err := s.repo.Update(client)
	if err != nil {
		return nil, fmt.Errorf("OAuthClientService: failed to update client: %w", err)
	}

	gl.Log("info", fmt.Sprintf("OAuthClientService: updated client %s", updated.GetClientID()))
	return updated, nil
}

func (s *OAuthClientService) DeactivateClient(id string) error {
	if id == "" {
		return fmt.Errorf("OAuthClientService: id is required")
	}

	client, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("OAuthClientService: failed to find client: %w", err)
	}

	client.SetActive(false)
	if _, err := s.repo.Update(client); err != nil {
		return fmt.Errorf("OAuthClientService: failed to deactivate client: %w", err)
	}

	gl.Log("info", fmt.Sprintf("OAuthClientService: deactivated client %s", client.GetClientID()))
	return nil
}

func (s *OAuthClientService) DeleteClient(id string) error {
	if id == "" {
		return fmt.Errorf("OAuthClientService: id is required")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("OAuthClientService: failed to delete client: %w", err)
	}

	gl.Log("info", fmt.Sprintf("OAuthClientService: deleted client with ID %s", id))
	return nil
}

func (s *OAuthClientService) ValidateRedirectURI(clientID, redirectURI string) error {
	if clientID == "" {
		return fmt.Errorf("OAuthClientService: client_id is required")
	}
	if redirectURI == "" {
		return fmt.Errorf("OAuthClientService: redirect_uri is required")
	}

	client, err := s.GetClientByClientID(clientID)
	if err != nil {
		return fmt.Errorf("OAuthClientService: failed to validate redirect_uri: %w", err)
	}

	if !client.IsRedirectURIAllowed(redirectURI) {
		return fmt.Errorf("OAuthClientService: redirect_uri not allowed for this client")
	}

	return nil
}
