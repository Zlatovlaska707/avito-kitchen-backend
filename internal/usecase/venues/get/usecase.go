package get

import (
	"context"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type UseCase struct {
	venues venuerepo.Repository
	nower  nowercontract.Nower
	defTZ  *time.Location
}

func New(venues venuerepo.Repository, nower nowercontract.Nower, defTZ *time.Location) *UseCase {
	return &UseCase{venues: venues, nower: nower, defTZ: defTZ}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	v, err := shared.ResolveVenue(ctx, uc.venues, in.VenueRef)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	if v.Status != "active" {
		return nil, usecase.ErrVenueNotFound
	}
	hours, err := uc.venues.GetWorkingHours(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	isOpen, reason := shared.VenueIsOpen(v, hours, uc.nower.Now(), uc.defTZ)
	var closedReason *string
	if !isOpen {
		closedReason = &reason
	}
	eta := v.AvgPrepMinutes + 10
	return &Out{Venue: *v, Hours: hours, IsOpen: isOpen, ClosedReason: closedReason, ETAMinutes: eta}, nil
}
