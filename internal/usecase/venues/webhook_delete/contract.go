package webhook_delete

import "github.com/google/uuid"

type In struct {
	VenueID   uuid.UUID
	WebhookID uuid.UUID
}

type Out struct{}
