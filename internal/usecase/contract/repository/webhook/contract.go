package webhook

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, in CreateIn) (*Webhook, error)
	Delete(ctx context.Context, venueID, webhookID uuid.UUID) error
	GetByID(ctx context.Context, venueID, webhookID uuid.UUID) (*Webhook, error)
	ListByVenue(ctx context.Context, venueID uuid.UUID) ([]Webhook, error)
	FindActiveByVenueAndEvent(ctx context.Context, venueID uuid.UUID, eventType string) ([]Webhook, error)
}

type Webhook struct {
	ID         uuid.UUID
	VenueID    uuid.UUID
	URL        string
	Secret     string
	EventTypes []string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateIn struct {
	VenueID    uuid.UUID
	URL        string
	Secret     string
	EventTypes []string
}
