// Package webhooks provides the webhook model and interface for the application
package webhooks

import (
	"net/url"
	"time"

	"github.com/google/uuid"
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	t "github.com/kubex-ecosystem/gdbase/internal/types"
	gl "github.com/kubex-ecosystem/logz"
)

type IWebhook interface {
	TableName() string
	GetID() uuid.UUID
	GetURL() string
	GetEvent() string
	GetStatus() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetURL(fullURL string) error
	SetEvent(event string)
	SetStatus(status string)
	SetCreatedAt(createdAt time.Time)
	SetUpdatedAt(updatedAt time.Time)
	SetWebhook(webhook Webhook)
	GetWebhook() Webhook
	IsValid() bool
	GetMapper() ci.IMapper[*Webhook]
	SetMapper(mapper ci.IMapper[*Webhook])
}

type Webhook struct {
	*t.Mutexes
	*t.Reference
	URL       *url.URL             `json:"url" xml:"url" yaml:"url" gorm:"column:url"`
	Event     string               `json:"event" xml:"event" yaml:"event" gorm:"column:event"`
	Status    string               `json:"status" xml:"status" yaml:"status" gorm:"column:status"`
	CreatedAt time.Time            `json:"created_at" xml:"created_at" yaml:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time            `json:"updated_at" xml:"updated_at" yaml:"updated_at" gorm:"column:updated_at"`
	Mapper    ci.IMapper[*Webhook] `json:"-" xml:"-" yaml:"-" gorm:"-"` // Não serializar
	// Você pode adicionar outros campos úteis para controle, como configurações do child server.
}

func NewWebhook(fullURL, event, status string) IWebhook {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		panic("Invalid URL: " + err.Error())
	}

	return &Webhook{
		Mutexes:   t.NewMutexesType(),
		Reference: t.NewReference(parsedURL.String()).GetReference(),
		URL:       parsedURL,
		Event:     event,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (w *Webhook) TableName() string {
	return "webhooks"
}
func (w *Webhook) GetID() uuid.UUID {
	return w.Reference.GetID()
}
func (w *Webhook) GetURL() string {
	return w.URL.String()
}
func (w *Webhook) GetEvent() string {
	return w.Event
}
func (w *Webhook) GetStatus() string {
	return w.Status
}
func (w *Webhook) GetCreatedAt() time.Time {
	return w.CreatedAt
}
func (w *Webhook) GetUpdatedAt() time.Time {
	return w.UpdatedAt
}
func (w *Webhook) SetURL(fullURL string) error {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return err
	}
	w.Reference.SetName(parsedURL.String())
	w.URL = parsedURL
	return nil
}
func (w *Webhook) SetEvent(event string) {
	w.Event = event
}
func (w *Webhook) SetStatus(status string) {
	w.Status = status
}
func (w *Webhook) SetCreatedAt(createdAt time.Time) {
	w.CreatedAt = createdAt
}
func (w *Webhook) SetUpdatedAt(updatedAt time.Time) {
	w.UpdatedAt = updatedAt
}
func (w *Webhook) SetWebhook(webhook Webhook) {
	w.Reference.ID = webhook.Reference.ID
	w.URL = webhook.URL
	w.Event = webhook.Event
	w.Status = webhook.Status
	w.CreatedAt = webhook.CreatedAt
	w.UpdatedAt = webhook.UpdatedAt
}
func (w *Webhook) GetWebhook() Webhook {
	return *w
}
func (w *Webhook) IsValid() bool {
	if w == nil {
		gl.Log("error", "WebhookModel: Webhook is nil")
		return false
	}
	if w.URL == nil {
		gl.Log("error", "WebhookModel: URL is nil")
		return false
	}
	if w.Event == "" {
		gl.Log("error", "WebhookModel: Event is empty")
		return false
	}
	if w.Status == "" {
		gl.Log("error", "WebhookModel: Status is empty")
		return false
	}
	if w.CreatedAt.IsZero() {
		gl.Log("error", "WebhookModel: CreatedAt is zero")
		return false
	}
	return true
}
func (w *Webhook) GetMapper() ci.IMapper[*Webhook] {
	if w.Mapper == nil {
		w.Mapper = t.NewMapper(&w, "")
	}
	return w.Mapper
}
func (w *Webhook) SetMapper(mapper ci.IMapper[*Webhook]) {
	if mapper == nil {
		gl.Log("error", "WebhookModel: mapper is nil")
		return
	}
	w.Mapper = mapper
}
