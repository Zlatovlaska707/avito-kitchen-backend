package get

import (
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type In struct {
	VenueRef string
}

type Out struct {
	Venue        venuerepo.Venue
	Hours        []venuerepo.WorkingHour
	IsOpen       bool
	ClosedReason *string
	ETAMinutes   int
}
