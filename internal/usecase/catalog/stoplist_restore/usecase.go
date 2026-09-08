package stoplist_restore

import (
	"context"

	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type UseCase struct {
	catalog catalogrepo.Repository
}

func New(catalog catalogrepo.Repository) *UseCase {
	return &UseCase{catalog: catalog}
}

func (uc *UseCase) Run(ctx context.Context, _ In) (*Out, error) {
	n, err := uc.catalog.RestoreExpiredStoplist(ctx)
	if err != nil {
		return nil, err
	}
	return &Out{Restored: n}, nil
}
