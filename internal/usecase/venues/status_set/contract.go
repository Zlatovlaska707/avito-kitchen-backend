package status_set

import "github.com/google/uuid"

type In struct {
	VenueID       uuid.UUID
	Status        *string
	AcceptsOrders *bool
	PauseReason   *string
}

type Out struct{}
