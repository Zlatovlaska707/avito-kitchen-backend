package status_set

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
	if err := uc.venues.UpdateStatus(ctx, venuerepo.StatusUpdateIn{
		VenueID: in.VenueID, Status: in.Status, AcceptsOrders: in.AcceptsOrders, PauseReason: in.PauseReason,
	}); err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	return &Out{}, nil
}
