package apikey_revoke

import "github.com/google/uuid"

type In struct {
	VenueID uuid.UUID
	KeyID   uuid.UUID
}

type Out struct{}
