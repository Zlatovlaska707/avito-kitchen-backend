package update

import (
	"github.com/google/uuid"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type In struct {
	ID               uuid.UUID
	Name             *string
	Description      *string
	AddressText      *string
	Lat              *float64
	Lon              *float64
	Timezone         *string
	AvgPrepMinutes   *int
	DeliveryFeeMinor *int64
	MinOrderMinor    *int64
	DeliveryRadiusM  *int
}

type Out struct {
	Venue venuerepo.Venue
}
