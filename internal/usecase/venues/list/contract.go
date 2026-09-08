package list

import (
	"time"

	"github.com/google/uuid"

	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type In struct {
	CityID   *int
	Lat      *float64
	Lon      *float64
	Search   *string
	OnlyOpen bool
	Sort     string
	Limit    int
	CursorTS *time.Time
	CursorID *uuid.UUID
}

type VenueOut struct {
	Venue      venuerepo.Venue
	IsOpen     bool
	ETAMinutes *int
	DistanceM  *int
}

type Out struct {
	Items   []VenueOut
	HasNext bool
	LastTS  *time.Time
	LastID  *uuid.UUID
}
