package apikey_issue

import "github.com/google/uuid"

type In struct {
	VenueID uuid.UUID
	Scopes  []string
}

type Out struct {
	KeyID  uuid.UUID
	APIKey string
	Prefix string
}
