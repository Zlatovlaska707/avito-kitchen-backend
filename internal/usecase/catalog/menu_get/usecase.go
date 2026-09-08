package menu_get

import (
	"context"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type UseCase struct {
	catalog catalogrepo.Repository
	venues  venuerepo.Repository
}

func New(catalog catalogrepo.Repository, venues venuerepo.Repository) *UseCase {
	return &UseCase{catalog: catalog, venues: venues}
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
	menu, etag, err := uc.catalog.GetMenu(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	return &Out{Menu: menu, ETag: etag}, nil
}
