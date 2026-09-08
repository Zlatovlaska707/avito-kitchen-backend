package onboard

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
	taken, err := uc.venues.IsSlugTaken(ctx, in.Slug, nil)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, usecase.ErrVenueSlugTaken
	}
	v, err := uc.venues.Create(ctx, venuerepo.CreateIn{
		Slug: in.Slug, Name: in.Name, Description: in.Description, CityID: in.CityID,
		AddressText: in.AddressText, Lat: in.Lat, Lon: in.Lon, Timezone: in.Timezone,
		AvgPrepMinutes: in.AvgPrepMinutes, DeliveryFeeMinor: in.DeliveryFeeMinor,
		MinOrderMinor: in.MinOrderMinor, DeliveryRadiusM: in.DeliveryRadiusM,
		WorkingHours: in.WorkingHours,
	})
	if err != nil {
		if pgerr.IsUniqueViolation(err) {
			return nil, usecase.ErrVenueSlugTaken
		}
		return nil, err
	}
	return &Out{Venue: *v}, nil
}
