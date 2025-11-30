// Package backend provides adapters for the kubex_be backend.
// This package implements the interfaces defined in kubex_be and bridges them
// to the actual database implementation (GORM + PostgreSQL).
package backend

import (
	// "context"
	"time"
	// Import kubexdb models
	// "github.com/kubex-ecosystem/gdbase/internal/models/invitations"
	// Import BE interface (the contract)
	// Note: In practice, this import path should match your actual module path
	// beInvitations "github.com/kubex-prm/kubex_be/internal/adapters/database/invitations"
)

// ===========================================================================
// TYPE MAPPINGS
// These types mirror the BE interface types to avoid circular imports.
// In production, you'd import them from kubex_be.
// ===========================================================================

type InvitationStatus string

const (
	StatusPending  InvitationStatus = "pending"
	StatusAccepted InvitationStatus = "accepted"
	StatusExpired  InvitationStatus = "expired"
	StatusRevoked  InvitationStatus = "revoked"
)

type InvitationType string

const (
	TypePartner  InvitationType = "partner"
	TypeInternal InvitationType = "internal"
)

type Invitation struct {
	ID         string           `json:"id"`
	Token      string           `json:"token"`
	Email      string           `json:"email"`
	Name       *string          `json:"name,omitempty"`
	Role       string           `json:"role"`
	Type       InvitationType   `json:"type"`
	Status     InvitationStatus `json:"status"`
	ExpiresAt  time.Time        `json:"expires_at"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
	TenantID   string           `json:"tenant_id"`
	InvitedBy  string           `json:"invited_by"`
	TeamID     *string          `json:"team_id,omitempty"`
	Company    *string          `json:"company,omitempty"`
	Metadata   *string          `json:"metadata,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  *time.Time       `json:"updated_at,omitempty"`
}

type CreatePartnerInvitationInput struct {
	PartnerEmail string     `json:"partner_email"`
	PartnerName  *string    `json:"partner_name,omitempty"`
	CompanyName  *string    `json:"company_name,omitempty"`
	Role         string     `json:"role"`
	TenantID     string     `json:"tenant_id"`
	InvitedBy    string     `json:"invited_by"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Metadata     *string    `json:"metadata,omitempty"`
}

type CreateInternalInvitationInput struct {
	InviteeEmail string     `json:"invitee_email"`
	InviteeName  *string    `json:"invitee_name,omitempty"`
	Role         string     `json:"role"`
	TeamID       *string    `json:"team_id,omitempty"`
	TenantID     string     `json:"tenant_id"`
	InvitedBy    string     `json:"invited_by"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Metadata     *string    `json:"metadata,omitempty"`
}

type UpdateInvitationInput struct {
	Status     *InvitationStatus `json:"status,omitempty"`
	AcceptedAt *time.Time        `json:"accepted_at,omitempty"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
	Metadata   *string           `json:"metadata,omitempty"`
}

type InvitationFilters struct {
	Email     *string           `json:"email,omitempty"`
	TenantID  *string           `json:"tenant_id,omitempty"`
	Status    *InvitationStatus `json:"status,omitempty"`
	InvitedBy *string           `json:"invited_by,omitempty"`
	Type      *InvitationType   `json:"type,omitempty"`
	Page      int               `json:"page"`
	Limit     int               `json:"limit"`
}

type PaginatedInvitations struct {
	Data       []*Invitation `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// ===========================================================================
// ADAPTER IMPLEMENTATION
// ===========================================================================

// InvitationsAdapter implements the BE Adapter interface using kubexdb
type InvitationsAdapter struct {
	// service invitations.Service
}

// NewInvitationsAdapter creates a new invitations adapter
// func NewInvitationsAdapter(service invitations.Service) *InvitationsAdapter {
// 	return &InvitationsAdapter{service: service}
// }

// // GetByToken retrieves an invitation by its token
// func (a *InvitationsAdapter) GetByToken(ctx context.Context, token string) (*Invitation, error) {
// 	genericInv, err := a.service.GetInvitationByToken(ctx, token)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapGenericToBE(genericInv), nil
// }

// // GetByID retrieves an invitation by its ID
// func (a *InvitationsAdapter) GetByID(ctx context.Context, id string, invType InvitationType) (*Invitation, error) {
// 	if invType == TypePartner {
// 		inv, err := a.service.GetPartnerInvitation(ctx, id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return mapPartnerToBE(inv), nil
// 	}

// 	inv, err := a.service.GetInternalInvitation(ctx, id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapInternalToBE(inv), nil
// }

// // CreatePartner creates a new partner invitation
// func (a *InvitationsAdapter) CreatePartner(ctx context.Context, input *CreatePartnerInvitationInput) (*Invitation, error) {
// 	dto := &invitations.CreatePartnerInvitationDTO{
// 		PartnerEmail: input.PartnerEmail,
// 		PartnerName:  input.PartnerName,
// 		CompanyName:  input.CompanyName,
// 		Role:         input.Role,
// 		TenantID:     input.TenantID,
// 		InvitedBy:    input.InvitedBy,
// 		ExpiresAt:    input.ExpiresAt,
// 		Metadata:     input.Metadata,
// 	}

// 	inv, err := a.service.CreatePartnerInvitation(ctx, dto)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapPartnerToBE(inv), nil
// }

// // CreateInternal creates a new internal invitation
// func (a *InvitationsAdapter) CreateInternal(ctx context.Context, input *CreateInternalInvitationInput) (*Invitation, error) {
// 	dto := &invitations.CreateInternalInvitationDTO{
// 		InviteeEmail: input.InviteeEmail,
// 		InviteeName:  input.InviteeName,
// 		Role:         input.Role,
// 		TeamID:       input.TeamID,
// 		TenantID:     input.TenantID,
// 		InvitedBy:    input.InvitedBy,
// 		ExpiresAt:    input.ExpiresAt,
// 		Metadata:     input.Metadata,
// 	}

// 	inv, err := a.service.CreateInternalInvitation(ctx, dto)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapInternalToBE(inv), nil
// }

// // Update updates an existing invitation
// func (a *InvitationsAdapter) Update(ctx context.Context, id string, invType InvitationType, input *UpdateInvitationInput) (*Invitation, error) {
// 	dto := &invitations.UpdateInvitationDTO{
// 		Status:     (*invitations.InvitationStatus)(input.Status),
// 		AcceptedAt: input.AcceptedAt,
// 		ExpiresAt:  input.ExpiresAt,
// 		Metadata:   input.Metadata,
// 	}

// 	if invType == TypePartner {
// 		inv, err := a.service.UpdatePartnerInvitation(ctx, id, dto)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return mapPartnerToBE(inv), nil
// 	}

// 	inv, err := a.service.UpdateInternalInvitation(ctx, id, dto)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapInternalToBE(inv), nil
// }

// // Revoke revokes an invitation
// func (a *InvitationsAdapter) Revoke(ctx context.Context, id string, invType InvitationType) error {
// 	if invType == TypePartner {
// 		return a.service.RevokePartnerInvitation(ctx, id)
// 	}
// 	return a.service.RevokeInternalInvitation(ctx, id)
// }

// // Accept accepts an invitation by token
// func (a *InvitationsAdapter) Accept(ctx context.Context, token string) (*Invitation, error) {
// 	// Try partner first
// 	partnerInv, err := a.service.AcceptPartnerInvitation(ctx, token)
// 	if err == nil {
// 		return mapPartnerToBE(partnerInv), nil
// 	}

// 	// Try internal
// 	internalInv, err := a.service.AcceptInternalInvitation(ctx, token)
// 	if err == nil {
// 		return mapInternalToBE(internalInv), nil
// 	}

// 	return nil, err
// }

// // Delete deletes an invitation
// func (a *InvitationsAdapter) Delete(ctx context.Context, id string, invType InvitationType) error {
// 	if invType == TypePartner {
// 		return a.service.DeletePartnerInvitation(ctx, id)
// 	}
// 	return a.service.DeleteInternalInvitation(ctx, id)
// }

// // List lists invitations with optional filtering
// func (a *InvitationsAdapter) List(ctx context.Context, filters *InvitationFilters) (*PaginatedInvitations, error) {
// 	dbFilters := &invitations.InvitationFilterParams{
// 		Email:     filters.Email,
// 		TenantID:  filters.TenantID,
// 		Status:    (*invitations.InvitationStatus)(filters.Status),
// 		InvitedBy: filters.InvitedBy,
// 		Type:      (*invitations.InvitationType)(filters.Type),
// 		Page:      filters.Page,
// 		Limit:     filters.Limit,
// 	}

// 	result, err := a.service.ListAllInvitations(ctx, dbFilters)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Map to BE types
// 	beInvitations := make([]*Invitation, len(result.Data))
// 	for i, inv := range result.Data {
// 		beInvitations[i] = mapGenericToBE(&inv)
// 	}

// 	return &PaginatedInvitations{
// 		Data:       beInvitations,
// 		Total:      result.Total,
// 		Page:       result.Page,
// 		Limit:      result.Limit,
// 		TotalPages: result.TotalPages,
// 	}, nil
// }

// // ===========================================================================
// // MAPPER FUNCTIONS
// // ===========================================================================

// // mapPartnerToBE converts a PartnerInvitation to BE Invitation
// func mapPartnerToBE(inv *invitations.PartnerInvitation) *Invitation {
// 	return &Invitation{
// 		ID:         inv.ID,
// 		Token:      inv.Token,
// 		Email:      inv.PartnerEmail,
// 		Name:       inv.PartnerName,
// 		Role:       inv.Role,
// 		Type:       TypePartner,
// 		Status:     InvitationStatus(inv.Status),
// 		ExpiresAt:  inv.ExpiresAt,
// 		AcceptedAt: inv.AcceptedAt,
// 		TenantID:   inv.TenantID,
// 		InvitedBy:  inv.InvitedBy,
// 		Company:    inv.CompanyName,
// 		Metadata:   inv.Metadata,
// 		CreatedAt:  inv.CreatedAt,
// 		UpdatedAt:  inv.UpdatedAt,
// 	}
// }

// // mapInternalToBE converts an InternalInvitation to BE Invitation
// func mapInternalToBE(inv *invitations.InternalInvitation) *Invitation {
// 	return &Invitation{
// 		ID:         inv.ID,
// 		Token:      inv.Token,
// 		Email:      inv.InviteeEmail,
// 		Name:       inv.InviteeName,
// 		Role:       inv.Role,
// 		Type:       TypeInternal,
// 		Status:     InvitationStatus(inv.Status),
// 		ExpiresAt:  inv.ExpiresAt,
// 		AcceptedAt: inv.AcceptedAt,
// 		TenantID:   inv.TenantID,
// 		InvitedBy:  inv.InvitedBy,
// 		TeamID:     inv.TeamID,
// 		Metadata:   inv.Metadata,
// 		CreatedAt:  inv.CreatedAt,
// 		UpdatedAt:  inv.UpdatedAt,
// 	}
// }

// // mapGenericToBE converts a GenericInvitation to BE Invitation
// func mapGenericToBE(inv *invitations.GenericInvitation) *Invitation {
// 	return &Invitation{
// 		ID:         inv.ID,
// 		Token:      inv.Token,
// 		Email:      inv.Email,
// 		Name:       inv.Name,
// 		Role:       inv.Role,
// 		Type:       InvitationType(inv.Type),
// 		Status:     InvitationStatus(inv.Status),
// 		ExpiresAt:  inv.ExpiresAt,
// 		AcceptedAt: inv.AcceptedAt,
// 		TenantID:   inv.TenantID,
// 		InvitedBy:  inv.InvitedBy,
// 		TeamID:     inv.TeamID,
// 		Company:    inv.Company,
// 		Metadata:   inv.Metadata,
// 		CreatedAt:  inv.CreatedAt,
// 		UpdatedAt:  inv.UpdatedAt,
// 	}
// }
