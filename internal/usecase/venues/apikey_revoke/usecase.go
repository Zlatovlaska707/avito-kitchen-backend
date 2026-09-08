package apikey_revoke

import (
	"context"

	apikeyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/apikey"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

type UseCase struct {
	apikeys *apikeyrepo.CachedRepository
}

func New(apikeys *apikeyrepo.CachedRepository) *UseCase {
	return &UseCase{apikeys: apikeys}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if _, err := uc.apikeys.Revoke(ctx, in.VenueID, in.KeyID); err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	return &Out{}, nil
}
