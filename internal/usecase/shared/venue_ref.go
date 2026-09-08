package shared

import (
	"context"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

func ResolveVenue(ctx context.Context, repo venuerepo.Repository, ref string) (*venuerepo.Venue, error) {
	if id, err := uuid.Parse(ref); err == nil {
		v, err := repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	v, err := repo.GetBySlug(ctx, ref)
	if err != nil {
		if pgerr.IsNotFound(err) {
			return nil, pgerr.ErrNotFound
		}
		return nil, err
	}
	return v, nil
}
