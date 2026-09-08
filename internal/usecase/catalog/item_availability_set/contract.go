package item_availability_set

import (
	"time"

	"github.com/google/uuid"
)

type In struct {
	VenueID          uuid.UUID
	ExternalID       string
	IsAvailable      bool
	UnavailableUntil *time.Time
}

type Out struct{}
