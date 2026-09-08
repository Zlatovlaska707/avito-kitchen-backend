package onboard

import venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"

type In struct {
	Slug             string
	Name             string
	Description      *string
	CityID           int
	AddressText      string
	Lat              float64
	Lon              float64
	Timezone         string
	AvgPrepMinutes   int
	DeliveryFeeMinor int64
	MinOrderMinor    int64
	DeliveryRadiusM  int
	WorkingHours     []venuerepo.WorkingHour
}

type Out struct {
	Venue venuerepo.Venue
}
