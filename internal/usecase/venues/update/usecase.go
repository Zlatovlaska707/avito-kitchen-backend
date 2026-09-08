package update

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type UseCase struct {
	venues venuerepo.Repository
}

func New(venues venuerepo.Repository) *UseCase {
	return &UseCase{venues: venues}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	v, err := uc.venues.Update(ctx, venuerepo.UpdateIn{
		ID: in.ID, Name: in.Name, Description: in.Description, AddressText: in.AddressText,
		Lat: in.Lat, Lon: in.Lon, Timezone: in.Timezone, AvgPrepMinutes: in.AvgPrepMinutes,
		DeliveryFeeMinor: in.DeliveryFeeMinor, MinOrderMinor: in.MinOrderMinor, DeliveryRadiusM: in.DeliveryRadiusM,
	})
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	return &Out{Venue: *v}, nil
}
