package item_availability_set

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type UseCase struct {
	catalog catalogrepo.Repository
}

func New(catalog catalogrepo.Repository) *UseCase {
	return &UseCase{catalog: catalog}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	err := uc.catalog.SetItemAvailability(ctx, catalogrepo.AvailabilityIn{
		VenueID: in.VenueID, ExternalID: in.ExternalID,
		IsAvailable: in.IsAvailable, UnavailableUntil: in.UnavailableUntil,
	})
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrMenuItemNotFound
		}
		return nil, err
	}
	return &Out{}, nil
}
